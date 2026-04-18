package kafka

import (
	"crypto/rand"
	"errors"
	"sort"
	"strings"

	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/internal/dal/response"

	"github.com/IBM/sarama"
)

func (s *Service) ListACLs(req reqKafka.ACLListRequest) ([]response.KafkaACLVO, error) {
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

	filter, err := buildACLFilter(req)
	if err != nil {
		return nil, err
	}
	resourceAcls, err := admin.ListAcls(filter)
	if err != nil {
		return nil, err
	}
	items := flattenResourceAcls(resourceAcls)
	sort.Slice(items, func(i, j int) bool {
		if items[i].ResourceType == items[j].ResourceType {
			if items[i].ResourceName == items[j].ResourceName {
				if items[i].Principal == items[j].Principal {
					if items[i].Operation == items[j].Operation {
						return items[i].Host < items[j].Host
					}
					return items[i].Operation < items[j].Operation
				}
				return items[i].Principal < items[j].Principal
			}
			return items[i].ResourceName < items[j].ResourceName
		}
		return items[i].ResourceType < items[j].ResourceType
	})
	return items, nil
}

func (s *Service) CreateACL(req reqKafka.ACLUpsertRequest) (*response.KafkaACLVO, error) {
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

	resourceAcls, err := buildACLResource(req.ResourceType, req.ResourceName, req.PatternType, req.Principal, req.Host, req.Operation, req.PermissionType)
	if err != nil {
		return nil, err
	}
	if err = admin.CreateACLs([]*sarama.ResourceAcls{resourceAcls}); err != nil {
		return nil, err
	}
	item := flattenResourceAcls([]sarama.ResourceAcls{*resourceAcls})
	if len(item) == 0 {
		return nil, errors.New("ACL 创建结果为空")
	}
	return &item[0], nil
}

func (s *Service) DeleteACL(req reqKafka.ACLDeleteRequest) (*response.KafkaACLDeleteResultVO, error) {
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

	resourceType, err := parseACLResourceType(req.ResourceType, true)
	if err != nil {
		return nil, err
	}
	patternType, err := parseACLPatternType(req.PatternType, false)
	if err != nil {
		return nil, err
	}
	operation, err := parseACLOperation(req.Operation, true)
	if err != nil {
		return nil, err
	}
	permissionType, err := parseACLPermissionType(req.PermissionType, true)
	if err != nil {
		return nil, err
	}
	resourceName := strings.TrimSpace(req.ResourceName)
	principal := strings.TrimSpace(req.Principal)
	host := strings.TrimSpace(req.Host)
	if host == "" {
		host = "*"
	}
	filter := sarama.AclFilter{
		Version:                   1,
		ResourceType:              resourceType,
		ResourceName:              &resourceName,
		ResourcePatternTypeFilter: patternType,
		Principal:                 &principal,
		Host:                      &host,
		Operation:                 operation,
		PermissionType:            permissionType,
	}
	matches, err := admin.DeleteACL(filter, false)
	if err != nil {
		return nil, err
	}
	deleted := make([]response.KafkaACLVO, 0, len(matches))
	for _, match := range matches {
		if match.Err != sarama.ErrNoError {
			continue
		}
		deleted = append(deleted, response.KafkaACLVO{
			ResourceType:   match.ResourceType.String(),
			ResourceName:   match.ResourceName,
			PatternType:    match.ResourcePatternType.String(),
			Principal:      match.Principal,
			Host:           match.Host,
			Operation:      match.Operation.String(),
			PermissionType: match.PermissionType.String(),
		})
	}
	return &response.KafkaACLDeleteResultVO{
		DeletedCount: len(deleted),
		Entries:      deleted,
	}, nil
}

func (s *Service) ListScramUsers(req reqKafka.ScramUserListRequest) ([]response.KafkaScramUserVO, error) {
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

	users, err := admin.DescribeUserScramCredentials(nil)
	if err != nil {
		return nil, err
	}
	keyword := strings.ToLower(strings.TrimSpace(req.Keyword))
	result := make([]response.KafkaScramUserVO, 0, len(users))
	for _, user := range users {
		if user == nil || user.ErrorCode != sarama.ErrNoError {
			continue
		}
		if keyword != "" && !strings.Contains(strings.ToLower(user.User), keyword) {
			continue
		}
		credentials := make([]response.KafkaScramCredentialVO, 0, len(user.CredentialInfos))
		for _, info := range user.CredentialInfos {
			if info == nil {
				continue
			}
			credentials = append(credentials, response.KafkaScramCredentialVO{
				Mechanism:  info.Mechanism.String(),
				Iterations: info.Iterations,
			})
		}
		sort.Slice(credentials, func(i, j int) bool { return credentials[i].Mechanism < credentials[j].Mechanism })
		result = append(result, response.KafkaScramUserVO{
			Username:    user.User,
			Credentials: credentials,
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Username < result[j].Username })
	return result, nil
}

func (s *Service) UpsertScramUser(req reqKafka.ScramUserUpsertRequest) error {
	cluster, err := s.repo.GetByID(req.ClusterID)
	if err != nil {
		return err
	}
	client, admin, err := s.openClients(cluster)
	if err != nil {
		return err
	}
	defer admin.Close()
	defer client.Close()

	mechanism, err := parseScramMechanism(req.Mechanism)
	if err != nil {
		return err
	}
	iterations := req.Iterations
	if iterations == 0 {
		iterations = 4096
	}
	salt := make([]byte, 32)
	if _, err = rand.Read(salt); err != nil {
		return err
	}
	results, err := admin.UpsertUserScramCredentials([]sarama.AlterUserScramCredentialsUpsert{
		{
			Name:       req.Username,
			Mechanism:  mechanism,
			Iterations: iterations,
			Salt:       salt,
			Password:   []byte(req.Password),
		},
	})
	if err != nil {
		return err
	}
	return ensureScramResultsOK(results)
}

func (s *Service) DeleteScramUser(req reqKafka.ScramUserDeleteRequest) error {
	cluster, err := s.repo.GetByID(req.ClusterID)
	if err != nil {
		return err
	}
	client, admin, err := s.openClients(cluster)
	if err != nil {
		return err
	}
	defer admin.Close()
	defer client.Close()

	mechanism, err := parseScramMechanism(req.Mechanism)
	if err != nil {
		return err
	}
	results, err := admin.DeleteUserScramCredentials([]sarama.AlterUserScramCredentialsDelete{
		{
			Name:      req.Username,
			Mechanism: mechanism,
		},
	})
	if err != nil {
		return err
	}
	return ensureScramResultsOK(results)
}

func buildACLFilter(req reqKafka.ACLListRequest) (sarama.AclFilter, error) {
	resourceType, err := parseACLResourceType(req.ResourceType, false)
	if err != nil {
		return sarama.AclFilter{}, err
	}
	patternType, err := parseACLPatternType(req.PatternType, true)
	if err != nil {
		return sarama.AclFilter{}, err
	}
	operation, err := parseACLOperation(req.Operation, false)
	if err != nil {
		return sarama.AclFilter{}, err
	}
	permissionType, err := parseACLPermissionType(req.PermissionType, false)
	if err != nil {
		return sarama.AclFilter{}, err
	}
	filter := sarama.AclFilter{
		Version:                   1,
		ResourceType:              resourceType,
		ResourcePatternTypeFilter: patternType,
		Operation:                 operation,
		PermissionType:            permissionType,
	}
	if value := strings.TrimSpace(req.ResourceName); value != "" {
		filter.ResourceName = &value
	}
	if value := strings.TrimSpace(req.Principal); value != "" {
		filter.Principal = &value
	}
	if value := strings.TrimSpace(req.Host); value != "" {
		filter.Host = &value
	}
	return filter, nil
}

func buildACLResource(resourceTypeRaw, resourceName, patternTypeRaw, principalRaw, hostRaw, operationRaw, permissionTypeRaw string) (*sarama.ResourceAcls, error) {
	resourceType, err := parseACLResourceType(resourceTypeRaw, true)
	if err != nil {
		return nil, err
	}
	patternType, err := parseACLPatternType(patternTypeRaw, false)
	if err != nil {
		return nil, err
	}
	operation, err := parseACLOperation(operationRaw, true)
	if err != nil {
		return nil, err
	}
	permissionType, err := parseACLPermissionType(permissionTypeRaw, true)
	if err != nil {
		return nil, err
	}
	host := strings.TrimSpace(hostRaw)
	if host == "" {
		host = "*"
	}
	return &sarama.ResourceAcls{
		Resource: sarama.Resource{
			ResourceType:        resourceType,
			ResourceName:        strings.TrimSpace(resourceName),
			ResourcePatternType: patternType,
		},
		Acls: []*sarama.Acl{
			{
				Principal:      strings.TrimSpace(principalRaw),
				Host:           host,
				Operation:      operation,
				PermissionType: permissionType,
			},
		},
	}, nil
}

func flattenResourceAcls(items []sarama.ResourceAcls) []response.KafkaACLVO {
	result := make([]response.KafkaACLVO, 0)
	for _, item := range items {
		for _, acl := range item.Acls {
			if acl == nil {
				continue
			}
			result = append(result, response.KafkaACLVO{
				ResourceType:   item.ResourceType.String(),
				ResourceName:   item.ResourceName,
				PatternType:    item.ResourcePatternType.String(),
				Principal:      acl.Principal,
				Host:           acl.Host,
				Operation:      acl.Operation.String(),
				PermissionType: acl.PermissionType.String(),
			})
		}
	}
	return result
}

func parseACLResourceType(raw string, required bool) (sarama.AclResourceType, error) {
	if strings.TrimSpace(raw) == "" {
		if required {
			return sarama.AclResourceUnknown, errors.New("请选择资源类型")
		}
		return sarama.AclResourceAny, nil
	}
	var value sarama.AclResourceType
	if err := value.UnmarshalText([]byte(raw)); err != nil {
		return sarama.AclResourceUnknown, errors.New("不支持的 ACL 资源类型")
	}
	return value, nil
}

func parseACLPatternType(raw string, allowAny bool) (sarama.AclResourcePatternType, error) {
	if strings.TrimSpace(raw) == "" {
		if allowAny {
			return sarama.AclPatternAny, nil
		}
		return sarama.AclPatternLiteral, nil
	}
	var value sarama.AclResourcePatternType
	if err := value.UnmarshalText([]byte(raw)); err != nil {
		return sarama.AclPatternUnknown, errors.New("不支持的 ACL Pattern 类型")
	}
	return value, nil
}

func parseACLOperation(raw string, required bool) (sarama.AclOperation, error) {
	if strings.TrimSpace(raw) == "" {
		if required {
			return sarama.AclOperationUnknown, errors.New("请选择 ACL 操作类型")
		}
		return sarama.AclOperationAny, nil
	}
	var value sarama.AclOperation
	if err := value.UnmarshalText([]byte(raw)); err != nil {
		return sarama.AclOperationUnknown, errors.New("不支持的 ACL 操作类型")
	}
	return value, nil
}

func parseACLPermissionType(raw string, required bool) (sarama.AclPermissionType, error) {
	if strings.TrimSpace(raw) == "" {
		if required {
			return sarama.AclPermissionUnknown, errors.New("请选择 ACL 权限类型")
		}
		return sarama.AclPermissionAny, nil
	}
	var value sarama.AclPermissionType
	if err := value.UnmarshalText([]byte(raw)); err != nil {
		return sarama.AclPermissionUnknown, errors.New("不支持的 ACL 权限类型")
	}
	return value, nil
}

func parseScramMechanism(raw string) (sarama.ScramMechanismType, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "sha256":
		return sarama.SCRAM_MECHANISM_SHA_256, nil
	case "sha512":
		return sarama.SCRAM_MECHANISM_SHA_512, nil
	default:
		return sarama.SCRAM_MECHANISM_SHA_256, errors.New("不支持的 SCRAM 机制")
	}
}

func ensureScramResultsOK(results []*sarama.AlterUserScramCredentialsResult) error {
	for _, result := range results {
		if result == nil {
			continue
		}
		if result.ErrorCode != sarama.ErrNoError {
			if result.ErrorMessage != nil && *result.ErrorMessage != "" {
				return errors.New(*result.ErrorMessage)
			}
			return result.ErrorCode
		}
	}
	return nil
}
