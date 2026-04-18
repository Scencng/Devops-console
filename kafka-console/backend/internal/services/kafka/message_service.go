package kafka

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/internal/dal/response"

	"github.com/IBM/sarama"
)

func (s *Service) BrowseMessages(req reqKafka.MessageBrowseRequest) (*response.KafkaMessageBrowseVO, error) {
	cluster, err := s.repo.GetByID(req.ClusterID)
	if err != nil {
		return nil, err
	}
	client, admin, err := s.openClients(cluster)
	if err != nil {
		return nil, err
	}
	defer admin.Close()
	defer client.Close()

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		return nil, err
	}
	defer consumer.Close()

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	mode := req.Mode
	if mode == "" {
		mode = "latest"
	}

	oldest, err := client.GetOffset(req.Topic, req.Partition, sarama.OffsetOldest)
	if err != nil {
		return nil, err
	}
	newest, err := client.GetOffset(req.Topic, req.Partition, sarama.OffsetNewest)
	if err != nil {
		return nil, err
	}

	startOffset := oldest
	switch mode {
	case "latest":
		startOffset = newest - int64(limit)
		if startOffset < oldest {
			startOffset = oldest
		}
	case "offset":
		startOffset = req.Offset
		if startOffset < oldest {
			startOffset = oldest
		}
		if startOffset > newest {
			startOffset = newest
		}
	case "earliest":
		startOffset = oldest
	}

	partitionConsumer, err := consumer.ConsumePartition(req.Topic, req.Partition, startOffset)
	if err != nil {
		return nil, err
	}
	defer partitionConsumer.Close()

	messages := make([]response.KafkaMessageVO, 0, limit)
	deadline := time.After(3 * time.Second)
	keyword := strings.ToLower(strings.TrimSpace(req.Keyword))

	for len(messages) < limit {
		select {
		case msg := <-partitionConsumer.Messages():
			if msg == nil {
				continue
			}
			keyPreview, keyBase64 := previewBytes(msg.Key)
			valuePreview, valueBase64 := previewBytes(msg.Value)
			if keyword != "" && !strings.Contains(strings.ToLower(keyPreview+" "+valuePreview), keyword) {
				continue
			}
			headers := make([]response.KafkaMessageHeaderVO, 0, len(msg.Headers))
			for _, header := range msg.Headers {
				headerValue, _ := previewBytes(header.Value)
				headers = append(headers, response.KafkaMessageHeaderVO{Key: string(header.Key), Value: headerValue})
			}
			messages = append(messages, response.KafkaMessageVO{Offset: msg.Offset, Partition: msg.Partition, Timestamp: msg.Timestamp, KeyPreview: keyPreview, ValuePreview: valuePreview, KeyBase64: keyBase64, ValueBase64: valueBase64, Headers: headers})
		case <-deadline:
			return &response.KafkaMessageBrowseVO{Topic: req.Topic, Partition: req.Partition, StartOffset: startOffset, Count: len(messages), Messages: messages}, nil
		case <-partitionConsumer.Errors():
			continue
		}
	}

	return &response.KafkaMessageBrowseVO{Topic: req.Topic, Partition: req.Partition, StartOffset: startOffset, Count: len(messages), Messages: messages}, nil
}

func (s *Service) ProduceMessage(req reqKafka.MessageProduceRequest) (*response.KafkaMessageProduceVO, error) {
	cluster, err := s.repo.GetByID(req.ClusterID)
	if err != nil {
		return nil, err
	}
	config, err := s.buildKafkaConfig(cluster)
	if err != nil {
		return nil, err
	}
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Idempotent = false
	if req.Partition != nil {
		config.Producer.Partitioner = sarama.NewManualPartitioner
	} else {
		config.Producer.Partitioner = sarama.NewHashPartitioner
	}

	producer, err := sarama.NewSyncProducer(normalizeBootstrapServers(cluster.BootstrapServers), config)
	if err != nil {
		return nil, err
	}
	defer producer.Close()

	valueBytes, err := decodeEncodedBytes(req.Value, req.ValueEncoding)
	if err != nil {
		return nil, err
	}
	keyBytes, err := decodeEncodedBytes(req.Key, req.KeyEncoding)
	if err != nil {
		return nil, err
	}

	message := &sarama.ProducerMessage{
		Topic:     req.Topic,
		Value:     sarama.ByteEncoder(valueBytes),
		Timestamp: time.Now(),
	}
	if len(keyBytes) > 0 {
		message.Key = sarama.ByteEncoder(keyBytes)
	}
	if req.Partition != nil {
		message.Partition = *req.Partition
	}
	if len(req.Headers) > 0 {
		headers := make([]sarama.RecordHeader, 0, len(req.Headers))
		for _, header := range req.Headers {
			headerValue, decodeErr := decodeEncodedBytes(header.Value, header.ValueEncoding)
			if decodeErr != nil {
				return nil, decodeErr
			}
			headers = append(headers, sarama.RecordHeader{
				Key:   []byte(header.Key),
				Value: headerValue,
			})
		}
		message.Headers = headers
	}

	partition, offset, err := producer.SendMessage(message)
	if err != nil {
		return nil, err
	}
	return &response.KafkaMessageProduceVO{
		Topic:     req.Topic,
		Partition: partition,
		Offset:    offset,
		Timestamp: message.Timestamp,
	}, nil
}

func previewBytes(data []byte) (string, string) {
	if len(data) == 0 {
		return "", ""
	}
	base64Value := base64.StdEncoding.EncodeToString(data)
	if utf8.Valid(data) {
		text := string(data)
		var pretty any
		if json.Unmarshal(data, &pretty) == nil {
			if prettyData, err := json.MarshalIndent(pretty, "", "  "); err == nil {
				text = string(prettyData)
			}
		}
		if len(text) > 1000 {
			text = text[:1000] + "..."
		}
		return text, base64Value
	}
	return base64Value, base64Value
}

func decodeEncodedBytes(raw, encoding string) ([]byte, error) {
	switch encoding {
	case "", "plain":
		return []byte(raw), nil
	case "base64":
		if strings.TrimSpace(raw) == "" {
			return nil, nil
		}
		value, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, errors.New("Base64 内容解析失败")
		}
		return value, nil
	default:
		return nil, errors.New("不支持的消息编码类型")
	}
}
