package kafka

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/internal/dal/response"

	"github.com/IBM/sarama"
)

type assignedGroupMember struct {
	memberID   string
	clientID   string
	clientHost string
}

// DeleteConsumerGroup only succeeds when the group is already in Empty state.
// Kafka does not allow deleting active consumer groups, so we fail fast before
// invoking sarama.ClusterAdmin.DeleteConsumerGroup.
func (s *Service) DeleteConsumerGroup(clusterID uint, groupID string) error {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return errors.New("消费组不能为空")
	}
	cluster, err := s.repo.GetByID(clusterID)
	if err != nil {
		return err
	}
	client, admin, err := s.openClients(cluster)
	if err != nil {
		return err
	}
	defer client.Close()
	defer admin.Close()

	groups, err := admin.DescribeConsumerGroups([]string{groupID})
	if err != nil {
		return err
	}
	if len(groups) == 0 || groups[0] == nil {
		return errors.New("消费组不存在")
	}
	group := groups[0]
	if len(group.Members) > 0 {
		return fmt.Errorf("消费组当前仍有 %d 个活跃成员（状态：%s），请先停止消费者后再删除", len(group.Members), group.State)
	}
	if !strings.EqualFold(strings.TrimSpace(group.State), "Empty") {
		return fmt.Errorf("消费组当前状态为 %s，仅 Empty 状态允许删除", group.State)
	}
	return admin.DeleteConsumerGroup(groupID)
}

func (s *Service) GetConsumerGroupDetail(groupID string, req reqKafka.ConsumerGroupDetailRequest) (*response.KafkaConsumerGroupDetailVO, error) {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, errors.New("消费组不能为空")
	}
	cluster, err := s.repo.GetByID(req.ClusterID)
	if err != nil {
		return nil, err
	}
	client, admin, err := s.openClients(cluster)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	defer admin.Close()

	groups, err := admin.DescribeConsumerGroups([]string{groupID})
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return nil, errors.New("消费组不存在")
	}
	group := groups[0]

	var topicFilter map[string][]int32
	if req.Topic != "" {
		topicFilter = map[string][]int32{
			req.Topic: nil,
		}
	}
	offsets, err := admin.ListConsumerGroupOffsets(groupID, topicFilter)
	if err != nil {
		return nil, err
	}

	assignmentMap := make(map[string]map[int32]assignedGroupMember)
	members := make([]response.KafkaConsumerGroupMemberVO, 0, len(group.Members))
	topicsSet := make(map[string]struct{})

	for _, member := range group.Members {
		memberTopicsSet := make(map[string]struct{})
		assignment, assignmentErr := member.GetMemberAssignment()
		if assignmentErr == nil && assignment != nil {
			for topic, partitions := range assignment.Topics {
				topicsSet[topic] = struct{}{}
				memberTopicsSet[topic] = struct{}{}
				if assignmentMap[topic] == nil {
					assignmentMap[topic] = make(map[int32]assignedGroupMember)
				}
				for _, partition := range partitions {
					assignmentMap[topic][partition] = assignedGroupMember{
						memberID:   member.MemberId,
						clientID:   member.ClientId,
						clientHost: member.ClientHost,
					}
				}
			}
		}

		memberTopics := make([]string, 0, len(memberTopicsSet))
		for topic := range memberTopicsSet {
			memberTopics = append(memberTopics, topic)
		}
		sort.Strings(memberTopics)
		members = append(members, response.KafkaConsumerGroupMemberVO{
			MemberID:   member.MemberId,
			ClientID:   member.ClientId,
			ClientHost: member.ClientHost,
			Topics:     memberTopics,
		})
	}

	sort.Slice(members, func(i, j int) bool { return members[i].MemberID < members[j].MemberID })

	partitions := make([]response.KafkaConsumerGroupPartitionLagVO, 0)
	var totalLag int64
	for topic, partitionBlocks := range offsets.Blocks {
		if req.Topic != "" && req.Topic != topic {
			continue
		}
		topicsSet[topic] = struct{}{}
		partitionIDs := make([]int32, 0, len(partitionBlocks))
		for partitionID := range partitionBlocks {
			partitionIDs = append(partitionIDs, partitionID)
		}
		sort.Slice(partitionIDs, func(i, j int) bool { return partitionIDs[i] < partitionIDs[j] })

		for _, partitionID := range partitionIDs {
			block := partitionBlocks[partitionID]
			committedOffset := int64(-1)
			if block != nil {
				committedOffset = block.Offset
			}
			oldestOffset, oldestErr := client.GetOffset(topic, partitionID, sarama.OffsetOldest)
			if oldestErr != nil {
				oldestOffset = -1
			}
			latestOffset, latestErr := client.GetOffset(topic, partitionID, sarama.OffsetNewest)
			if latestErr != nil {
				latestOffset = -1
			}
			lag := int64(-1)
			if committedOffset >= 0 && latestOffset >= 0 {
				lag = 0
				if latestOffset > committedOffset {
					lag = latestOffset - committedOffset
				}
				totalLag += lag
			}
			assigned := assignmentMap[topic][partitionID]
			partitions = append(partitions, response.KafkaConsumerGroupPartitionLagVO{
				Topic:           topic,
				Partition:       partitionID,
				CommittedOffset: committedOffset,
				LatestOffset:    latestOffset,
				OldestOffset:    oldestOffset,
				Lag:             lag,
				MemberID:        assigned.memberID,
				ClientID:        assigned.clientID,
				ClientHost:      assigned.clientHost,
			})
		}
	}

	sort.Slice(partitions, func(i, j int) bool {
		if partitions[i].Topic == partitions[j].Topic {
			return partitions[i].Partition < partitions[j].Partition
		}
		return partitions[i].Topic < partitions[j].Topic
	})

	topics := make([]string, 0, len(topicsSet))
	for topic := range topicsSet {
		topics = append(topics, topic)
	}
	sort.Strings(topics)

	return &response.KafkaConsumerGroupDetailVO{
		GroupID:        group.GroupId,
		ProtocolType:   group.ProtocolType,
		State:          group.State,
		MemberCount:    len(group.Members),
		PartitionCount: len(partitions),
		TotalLag:       totalLag,
		Topics:         topics,
		Members:        members,
		Partitions:     partitions,
	}, nil
}
