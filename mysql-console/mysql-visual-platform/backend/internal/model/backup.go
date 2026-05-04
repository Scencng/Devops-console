package model

type BackupScope string

const (
	BackupScopeDatabase BackupScope = "database"
	BackupScopeTable    BackupScope = "table"
)

type BackupRecord struct {
	ID          string      `json:"id"`
	Database    string      `json:"database"`
	TableName   string      `json:"tableName,omitempty"`
	Scope       BackupScope `json:"scope"`
	FileName    string      `json:"fileName"`
	DisplayName string      `json:"displayName"`
	Size        int64       `json:"size"`
	Compressed  bool        `json:"compressed"`
	CreatedAt   string      `json:"createdAt"`
	UpdatedAt   string      `json:"updatedAt"`
}

type BackupListResponse struct {
	Records []BackupRecord `json:"records"`
}

type CreateBackupRequest struct {
	Database  string `json:"database" binding:"required"`
	TableName string `json:"tableName"`
	Compress  bool   `json:"compress"`
}

type CreateBackupResponse struct {
	TaskID string `json:"taskId"`
}

type RestoreBackupRequest struct {
	Database       string `json:"database" binding:"required"`
	FileName       string `json:"fileName" binding:"required"`
	TargetDatabase string `json:"targetDatabase" binding:"required"`
}

type RenameBackupRequest struct {
	Database string `json:"database" binding:"required"`
	FileName string `json:"fileName" binding:"required"`
	NewName  string `json:"newName" binding:"required"`
}

type DeleteBackupRequest struct {
	Database string `json:"database" binding:"required"`
	FileName string `json:"fileName" binding:"required"`
}

type BackupTaskStatus string

const (
	BackupTaskPending BackupTaskStatus = "pending"
	BackupTaskRunning BackupTaskStatus = "running"
	BackupTaskSuccess BackupTaskStatus = "success"
	BackupTaskFailed  BackupTaskStatus = "failed"
)

type BackupTask struct {
	ID             string           `json:"id"`
	Type           string           `json:"type"`
	Database       string           `json:"database"`
	TargetDatabase string           `json:"targetDatabase,omitempty"`
	TableName      string           `json:"tableName,omitempty"`
	FileName       string           `json:"fileName,omitempty"`
	Status         BackupTaskStatus `json:"status"`
	Progress       int              `json:"progress"`
	Message        string           `json:"message,omitempty"`
	CreatedAt      string           `json:"createdAt"`
	UpdatedAt      string           `json:"updatedAt"`
	CompletedAt    string           `json:"completedAt,omitempty"`
}

type CreateBackupScheduleRequest struct {
	Database        string `json:"database" binding:"required"`
	TableName       string `json:"tableName"`
	IntervalMinutes int    `json:"intervalMinutes" binding:"required,min=1"`
	Compress        bool   `json:"compress"`
}

type DeleteBackupScheduleRequest struct {
	ID string `json:"id" binding:"required"`
}

type BackupSchedule struct {
	ID              string `json:"id"`
	Database        string `json:"database"`
	TableName       string `json:"tableName,omitempty"`
	IntervalMinutes int    `json:"intervalMinutes"`
	Compress        bool   `json:"compress"`
	NextRunAt       string `json:"nextRunAt"`
	LastRunAt       string `json:"lastRunAt,omitempty"`
	LastStatus      string `json:"lastStatus,omitempty"`
}
