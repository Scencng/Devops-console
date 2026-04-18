package kafka

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"devops-console-backend/internal/dal"
	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/internal/dal/response"
	"devops-console-backend/pkg/configs"
)

func (s *Service) ListTasks(req reqKafka.TaskListRequest) ([]response.KafkaTaskVO, error) {
	list, err := configs.NewKafkaTaskRepository().List(req.ClusterID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaTaskVO, 0, len(list))
	for _, item := range list {
		result = append(result, toTaskVO(item))
	}
	return result, nil
}

func (s *Service) CreateTask(req reqKafka.TaskUpsertRequest) (*response.KafkaTaskVO, error) {
	item := &dal.KafkaTask{
		ClusterID:   req.ClusterID,
		Name:        req.Name,
		TaskType:    req.TaskType,
		Payload:     req.Payload,
		CronExpr:    req.CronExpr,
		Enabled:     req.Enabled,
		Environment: req.Environment,
		Tenant:      req.Tenant,
	}
	if err := configs.NewKafkaTaskRepository().Create(item); err != nil {
		return nil, err
	}
	vo := toTaskVO(*item)
	return &vo, nil
}

func (s *Service) UpdateTask(id uint, req reqKafka.TaskUpsertRequest) (*response.KafkaTaskVO, error) {
	repo := configs.NewKafkaTaskRepository()
	item, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	item.ClusterID = req.ClusterID
	item.Name = req.Name
	item.TaskType = req.TaskType
	item.Payload = req.Payload
	item.CronExpr = req.CronExpr
	item.Enabled = req.Enabled
	item.Environment = req.Environment
	item.Tenant = req.Tenant
	if err = repo.Update(item); err != nil {
		return nil, err
	}
	vo := toTaskVO(*item)
	return &vo, nil
}

func (s *Service) DeleteTask(id uint) error {
	return configs.NewKafkaTaskRepository().Delete(id)
}

func (s *Service) RunTask(id uint, triggerMode string, operator string) (*response.KafkaTaskRunVO, error) {
	repo := configs.NewKafkaTaskRepository()
	task, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	startedAt := time.Now()
	run := &dal.KafkaTaskRun{
		TaskID:      task.ID,
		ClusterID:   task.ClusterID,
		Status:      "running",
		TriggerMode: triggerMode,
		StartedAt:   startedAt,
	}
	if run.TriggerMode == "" {
		run.TriggerMode = "manual"
	}
	if err = repo.CreateRun(run); err != nil {
		return nil, err
	}

	status := "success"
	summary := fmt.Sprintf("任务由 %s 触发", operator)
	payloadResult := ""
	if execErr := s.executeTask(task.TaskType, task.Payload, task.ClusterID, operator); execErr != nil {
		status = "failed"
		summary = execErr.Error()
		payloadResult = execErr.Error()
	}
	finishedAt := time.Now()
	run.Status = status
	run.ResultSummary = summary
	run.ResultPayload = payloadResult
	run.FinishedAt = &finishedAt
	_ = repo.UpdateRun(run)
	task.LastRunStatus = status
	_ = repo.Update(task)
	vo := toTaskRunVO(*run)
	return &vo, nil
}

func (s *Service) ListTaskRuns(req reqKafka.TaskRunListRequest) ([]response.KafkaTaskRunVO, error) {
	list, err := configs.NewKafkaTaskRepository().ListRuns(req.TaskID)
	if err != nil {
		return nil, err
	}
	result := make([]response.KafkaTaskRunVO, 0, len(list))
	for _, item := range list {
		result = append(result, toTaskRunVO(item))
	}
	return result, nil
}

func (s *Service) executeTask(taskType, payload string, clusterID uint, operator string) error {
	switch taskType {
	case "inspection.run":
		_, err := s.RunInspection(reqKafka.InspectionRunRequest{ClusterID: clusterID, Name: "自动巡检任务"}, operator)
		return err
	case "alerts.evaluate":
		_, err := s.EvaluateAlertRules(clusterID)
		return err
	case "topic.create":
		var req reqKafka.CreateTopicRequest
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			return err
		}
		req.ClusterID = clusterID
		_, err := s.CreateTopic(req)
		return err
	case "topic.partitions.increase":
		var req struct {
			Topic string `json:"topic"`
			Count int32  `json:"count"`
		}
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			return err
		}
		_, err := s.IncreaseTopicPartitions(req.Topic, reqKafka.IncreaseTopicPartitionsRequest{ClusterID: clusterID, Count: req.Count})
		return err
	case "group.offset.reset":
		var req struct {
			GroupID string                              `json:"groupId"`
			Data    reqKafka.ResetConsumerGroupOffsetRequest `json:"data"`
		}
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			return err
		}
		req.Data.ClusterID = clusterID
		_, err := s.ResetConsumerGroupOffset(req.GroupID, req.Data)
		return err
	default:
		return errors.New("不支持的任务类型")
	}
}

func toTaskVO(item dal.KafkaTask) response.KafkaTaskVO {
	return response.KafkaTaskVO{
		ID:            item.ID,
		ClusterID:     item.ClusterID,
		Name:          item.Name,
		TaskType:      item.TaskType,
		Payload:       item.Payload,
		CronExpr:      item.CronExpr,
		Enabled:       item.Enabled,
		LastRunStatus: item.LastRunStatus,
		Environment:   item.Environment,
		Tenant:        item.Tenant,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func toTaskRunVO(item dal.KafkaTaskRun) response.KafkaTaskRunVO {
	return response.KafkaTaskRunVO{
		ID:            item.ID,
		TaskID:        item.TaskID,
		ClusterID:     item.ClusterID,
		Status:        item.Status,
		TriggerMode:   item.TriggerMode,
		ResultSummary: item.ResultSummary,
		ResultPayload: item.ResultPayload,
		StartedAt:     item.StartedAt,
		FinishedAt:    item.FinishedAt,
	}
}
