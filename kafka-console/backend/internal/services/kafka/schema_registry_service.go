package kafka

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"devops-console-backend/internal/dal"
	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/internal/dal/response"
	"devops-console-backend/pkg/configs"
	cryptoutil "devops-console-backend/pkg/utils/crypto"
)

type schemaVersionResponse struct {
	Subject     string      `json:"subject"`
	Version     int         `json:"version"`
	ID          int         `json:"id"`
	SchemaType  string      `json:"schemaType"`
	Schema      string      `json:"schema"`
	References  interface{} `json:"references"`
}

type schemaCompatibilityResponse struct {
	IsCompatible bool `json:"is_compatible"`
}

func (s *Service) ListSchemaRegistries(req reqKafka.SchemaRegistryListRequest) ([]response.KafkaSchemaRegistryVO, error) {
	list, err := configs.NewKafkaSchemaRegistryRepository().List(req.ClusterID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaSchemaRegistryVO, 0, len(list))
	for _, item := range list {
		result = append(result, toSchemaRegistryVO(item))
	}
	return result, nil
}

func (s *Service) CreateSchemaRegistry(req reqKafka.SchemaRegistryUpsertRequest) (*response.KafkaSchemaRegistryVO, error) {
	item := &dal.KafkaSchemaRegistry{
		ClusterID:   req.ClusterID,
		Name:        req.Name,
		Endpoint:    strings.TrimRight(req.Endpoint, "/"),
		AuthType:    normalizeRegistryAuthType(req.AuthType),
		Username:    req.Username,
		VerifySSL:   req.VerifySSL,
		Environment: req.Environment,
		Tenant:      req.Tenant,
		Description: req.Description,
	}
	if req.Password != "" {
		cipherText, err := cryptoutil.EncryptString(req.Password)
		if err != nil {
			return nil, err
		}
		item.PasswordCiphertext = cipherText
	}
	if err := configs.NewKafkaSchemaRegistryRepository().Create(item); err != nil {
		return nil, err
	}
	vo := toSchemaRegistryVO(*item)
	return &vo, nil
}

func (s *Service) UpdateSchemaRegistry(id uint, req reqKafka.SchemaRegistryUpsertRequest) (*response.KafkaSchemaRegistryVO, error) {
	repo := configs.NewKafkaSchemaRegistryRepository()
	item, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	item.ClusterID = req.ClusterID
	item.Name = req.Name
	item.Endpoint = strings.TrimRight(req.Endpoint, "/")
	item.AuthType = normalizeRegistryAuthType(req.AuthType)
	item.Username = req.Username
	item.VerifySSL = req.VerifySSL
	item.Environment = req.Environment
	item.Tenant = req.Tenant
	item.Description = req.Description
	if req.Password != "" {
		cipherText, cipherErr := cryptoutil.EncryptString(req.Password)
		if cipherErr != nil {
			return nil, cipherErr
		}
		item.PasswordCiphertext = cipherText
	}
	if err = repo.Update(item); err != nil {
		return nil, err
	}
	vo := toSchemaRegistryVO(*item)
	return &vo, nil
}

func (s *Service) DeleteSchemaRegistry(id uint) error {
	return configs.NewKafkaSchemaRegistryRepository().Delete(id)
}

func (s *Service) ListSchemaSubjects(req reqKafka.SchemaSubjectListRequest) ([]response.KafkaSchemaSubjectVO, error) {
	registry, err := s.getActiveSchemaRegistry(req.ClusterID)
	if err != nil {
		return nil, err
	}
	var subjects []string
	if err = s.schemaRegistryRequest(registry, http.MethodGet, "/subjects", nil, &subjects); err != nil {
		return nil, err
	}
	result := make([]response.KafkaSchemaSubjectVO, 0, len(subjects))
	for _, item := range subjects {
		result = append(result, response.KafkaSchemaSubjectVO{Subject: item})
	}
	return result, nil
}

func (s *Service) ListSchemaVersions(req reqKafka.SchemaVersionListRequest) ([]response.KafkaSchemaVersionVO, error) {
	registry, err := s.getActiveSchemaRegistry(req.ClusterID)
	if err != nil {
		return nil, err
	}
	var versions []int
	if err = s.schemaRegistryRequest(registry, http.MethodGet, fmt.Sprintf("/subjects/%s/versions", req.Subject), nil, &versions); err != nil {
		return nil, err
	}
	result := make([]response.KafkaSchemaVersionVO, 0, len(versions))
	for _, item := range versions {
		result = append(result, response.KafkaSchemaVersionVO{Version: item})
	}
	return result, nil
}

func (s *Service) GetSchemaDetail(req reqKafka.SchemaDetailRequest) (*response.KafkaSchemaDetailVO, error) {
	registry, err := s.getActiveSchemaRegistry(req.ClusterID)
	if err != nil {
		return nil, err
	}
	var detail schemaVersionResponse
	if err = s.schemaRegistryRequest(registry, http.MethodGet, fmt.Sprintf("/subjects/%s/versions/%s", req.Subject, req.Version), nil, &detail); err != nil {
		return nil, err
	}
	vo := &response.KafkaSchemaDetailVO{
		Subject:    detail.Subject,
		Version:    detail.Version,
		ID:         detail.ID,
		SchemaType: detail.SchemaType,
		Schema:     detail.Schema,
		References: detail.References,
	}
	return vo, nil
}

func (s *Service) CheckSchemaCompatibility(req reqKafka.SchemaCompatibilityRequest) (*response.KafkaSchemaCompatibilityVO, error) {
	registry, err := s.getActiveSchemaRegistry(req.ClusterID)
	if err != nil {
		return nil, err
	}
	payload := map[string]string{"schema": req.Schema}
	var result schemaCompatibilityResponse
	if err = s.schemaRegistryRequest(registry, http.MethodPost, fmt.Sprintf("/compatibility/subjects/%s/versions/%s", req.Subject, req.Version), payload, &result); err != nil {
		return nil, err
	}
	message := "兼容"
	if !result.IsCompatible {
		message = "不兼容"
	}
	return &response.KafkaSchemaCompatibilityVO{IsCompatible: result.IsCompatible, Message: message}, nil
}

func (s *Service) getActiveSchemaRegistry(clusterID uint) (*dal.KafkaSchemaRegistry, error) {
	registry, err := configs.NewKafkaSchemaRegistryRepository().GetFirstByCluster(clusterID)
	if err != nil {
		return nil, errors.New("请先配置 Schema Registry")
	}
	return registry, nil
}

func (s *Service) schemaRegistryRequest(registry *dal.KafkaSchemaRegistry, method, path string, payload interface{}, target interface{}) error {
	url := strings.TrimRight(registry.Endpoint, "/") + path
	var body io.Reader
	if payload != nil {
		bytesPayload, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(bytesPayload)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/vnd.schemaregistry.v1+json")
	if registry.AuthType == "basic" && registry.Username != "" && registry.PasswordCiphertext != "" {
		password, err := cryptoutil.DecryptString(registry.PasswordCiphertext)
		if err != nil {
			return err
		}
		req.SetBasicAuth(registry.Username, password)
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Schema Registry 请求失败: HTTP %d - %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}
	if target == nil {
		return nil
	}
	return json.Unmarshal(bodyBytes, target)
}

func normalizeRegistryAuthType(value string) string {
	switch value {
	case "basic":
		return value
	default:
		return "none"
	}
}

func toSchemaRegistryVO(item dal.KafkaSchemaRegistry) response.KafkaSchemaRegistryVO {
	return response.KafkaSchemaRegistryVO{
		ID:          item.ID,
		ClusterID:   item.ClusterID,
		Name:        item.Name,
		Endpoint:    item.Endpoint,
		AuthType:    item.AuthType,
		Username:    item.Username,
		VerifySSL:   item.VerifySSL,
		Environment: item.Environment,
		Tenant:      item.Tenant,
		Description: item.Description,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}
