package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"devops-console-backend/internal/mysql/model"
)

type BackupService struct {
	baseDir   string
	tasksMu   sync.RWMutex
	tasks     map[string]*model.BackupTask
	schedMu   sync.RWMutex
	schedules map[string]*backupScheduleRuntime
}

type backupMetadata struct {
	ID          string            `json:"id"`
	Database    string            `json:"database"`
	TableName   string            `json:"tableName,omitempty"`
	Scope       model.BackupScope `json:"scope"`
	FileName    string            `json:"fileName"`
	DisplayName string            `json:"displayName"`
	Compressed  bool              `json:"compressed"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type backupScheduleRuntime struct {
	schedule model.BackupSchedule
	cancel   context.CancelFunc
}

func NewBackupService(baseDir string) *BackupService {
	return &BackupService{
		baseDir:   baseDir,
		tasks:     make(map[string]*model.BackupTask),
		schedules: make(map[string]*backupScheduleRuntime),
	}
}

func (s *BackupService) ListBackups(database string) ([]model.BackupRecord, error) {
	dir, err := s.ensureDatabaseDir(database)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	records := make([]model.BackupRecord, 0)
	for _, entry := range entries {
		if entry.IsDir() || strings.HasSuffix(entry.Name(), ".meta.json") {
			continue
		}

		record, err := s.loadBackupRecord(database, entry.Name())
		if err != nil {
			continue
		}

		records = append(records, *record)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].UpdatedAt > records[j].UpdatedAt
	})
	return records, nil
}

func (s *BackupService) CreateBackupAsync(ctx context.Context, profile model.OpenConnectionRequest, req model.CreateBackupRequest) (string, error) {
	task := s.newTask("backup", req.Database, "", req.TableName, "")
	go s.runBackupTask(context.Background(), task.ID, profile, req)
	return task.ID, nil
}

func (s *BackupService) RestoreBackupAsync(ctx context.Context, profile model.OpenConnectionRequest, req model.RestoreBackupRequest) (string, error) {
	task := s.newTask("restore", req.Database, req.TargetDatabase, "", req.FileName)
	go s.runRestoreTask(context.Background(), task.ID, profile, req)
	return task.ID, nil
}

func (s *BackupService) GetTask(id string) (*model.BackupTask, error) {
	s.tasksMu.RLock()
	defer s.tasksMu.RUnlock()
	task, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("backup task not found")
	}

	cloned := *task
	return &cloned, nil
}

func (s *BackupService) RenameBackup(req model.RenameBackupRequest) (*model.BackupRecord, error) {
	record, err := s.loadBackupRecord(req.Database, req.FileName)
	if err != nil {
		return nil, err
	}

	displayName := sanitizeBackupDisplayName(req.NewName)
	if displayName == "" {
		return nil, fmt.Errorf("new backup name is required")
	}

	dir, err := s.ensureDatabaseDir(req.Database)
	if err != nil {
		return nil, err
	}

	nextFileName := buildRenamedBackupFileName(displayName, req.FileName)
	oldPath := filepath.Join(dir, req.FileName)
	newPath := filepath.Join(dir, nextFileName)
	if oldPath == newPath {
		record.DisplayName = displayName
		return s.saveBackupMetadata(req.Database, nextFileName, recordToMetadata(*record))
	}

	if _, err := os.Stat(newPath); err == nil {
		return nil, fmt.Errorf("backup file %s already exists", nextFileName)
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return nil, err
	}

	oldMeta := s.metadataPath(req.Database, req.FileName)
	newMeta := s.metadataPath(req.Database, nextFileName)
	_ = os.Remove(newMeta)
	if _, err := os.Stat(oldMeta); err == nil {
		if err := os.Rename(oldMeta, newMeta); err != nil {
			return nil, err
		}
	}

	record.FileName = nextFileName
	record.DisplayName = displayName
	record.UpdatedAt = time.Now().Format(time.RFC3339)
	return s.saveBackupMetadata(req.Database, nextFileName, recordToMetadata(*record))
}

func (s *BackupService) DeleteBackup(req model.DeleteBackupRequest) error {
	record, err := s.loadBackupRecord(req.Database, req.FileName)
	if err != nil {
		return err
	}

	dir, err := s.ensureDatabaseDir(req.Database)
	if err != nil {
		return err
	}

	if err := os.Remove(filepath.Join(dir, record.FileName)); err != nil && !os.IsNotExist(err) {
		return err
	}

	if err := os.Remove(s.metadataPath(req.Database, record.FileName)); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (s *BackupService) OpenBackupReadCloser(database, fileName string) (io.ReadCloser, *model.BackupRecord, error) {
	record, err := s.loadBackupRecord(database, fileName)
	if err != nil {
		return nil, nil, err
	}

	dir, err := s.ensureDatabaseDir(database)
	if err != nil {
		return nil, nil, err
	}

	reader, err := os.Open(filepath.Join(dir, record.FileName))
	if err != nil {
		return nil, nil, err
	}

	return reader, record, nil
}

func (s *BackupService) ListSchedules(database string) []model.BackupSchedule {
	s.schedMu.RLock()
	defer s.schedMu.RUnlock()

	result := make([]model.BackupSchedule, 0)
	for _, runtimeSchedule := range s.schedules {
		if runtimeSchedule.schedule.Database == database {
			result = append(result, runtimeSchedule.schedule)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].NextRunAt < result[j].NextRunAt
	})
	return result
}

func (s *BackupService) CreateSchedule(profile model.OpenConnectionRequest, req model.CreateBackupScheduleRequest) (*model.BackupSchedule, error) {
	id := uuid.NewString()
	now := time.Now()
	schedule := model.BackupSchedule{
		ID:              id,
		Database:        req.Database,
		TableName:       strings.TrimSpace(req.TableName),
		IntervalMinutes: req.IntervalMinutes,
		Compress:        req.Compress,
		NextRunAt:       now.Add(time.Duration(req.IntervalMinutes) * time.Minute).Format(time.RFC3339),
	}

	ctx, cancel := context.WithCancel(context.Background())
	runtimeSchedule := &backupScheduleRuntime{
		schedule: schedule,
		cancel:   cancel,
	}

	s.schedMu.Lock()
	s.schedules[id] = runtimeSchedule
	s.schedMu.Unlock()

	go s.runSchedule(ctx, id, profile, req)
	return &schedule, nil
}

func (s *BackupService) DeleteSchedule(id string) error {
	s.schedMu.Lock()
	defer s.schedMu.Unlock()
	runtimeSchedule, ok := s.schedules[id]
	if !ok {
		return fmt.Errorf("backup schedule not found")
	}
	runtimeSchedule.cancel()
	delete(s.schedules, id)
	return nil
}

func (s *BackupService) runSchedule(ctx context.Context, id string, profile model.OpenConnectionRequest, req model.CreateBackupScheduleRequest) {
	ticker := time.NewTicker(time.Duration(req.IntervalMinutes) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			task := s.newTask("backup", req.Database, "", req.TableName, "")
			s.runBackupTask(context.Background(), task.ID, profile, model.CreateBackupRequest{
				Database:  req.Database,
				TableName: req.TableName,
				Compress:  req.Compress,
			})

			s.schedMu.Lock()
			runtimeSchedule, ok := s.schedules[id]
			if ok {
				now := time.Now()
				runtimeSchedule.schedule.LastRunAt = now.Format(time.RFC3339)
				if taskState, taskErr := s.GetTask(task.ID); taskErr == nil {
					runtimeSchedule.schedule.LastStatus = string(taskState.Status)
				}
				runtimeSchedule.schedule.NextRunAt = now.Add(time.Duration(req.IntervalMinutes) * time.Minute).Format(time.RFC3339)
			}
			s.schedMu.Unlock()
		}
	}
}

func (s *BackupService) runBackupTask(ctx context.Context, taskID string, profile model.OpenConnectionRequest, req model.CreateBackupRequest) {
	s.updateTask(taskID, model.BackupTaskRunning, 10, "resolving mysqldump")

	mysqldumpBin, err := resolveMySQLExecutable("mysqldump")
	if err != nil {
		s.failTask(taskID, err)
		return
	}

	record, dumpPath, err := s.prepareBackupTarget(req)
	if err != nil {
		s.failTask(taskID, err)
		return
	}

	s.updateTask(taskID, model.BackupTaskRunning, 30, "dumping database")
	if err := s.executeDump(ctx, mysqldumpBin, profile, req, dumpPath); err != nil {
		s.failTask(taskID, err)
		_ = os.Remove(dumpPath)
		return
	}

	s.updateTask(taskID, model.BackupTaskRunning, 85, "writing metadata")
	if _, err := s.saveBackupMetadata(req.Database, record.FileName, record); err != nil {
		s.failTask(taskID, err)
		return
	}

	s.finishTask(taskID, record.FileName)
}

func (s *BackupService) runRestoreTask(ctx context.Context, taskID string, profile model.OpenConnectionRequest, req model.RestoreBackupRequest) {
	mysqlBin, err := resolveMySQLExecutable("mysql")
	if err != nil {
		s.failTask(taskID, err)
		return
	}

	record, err := s.loadBackupRecord(req.Database, req.FileName)
	if err != nil {
		s.failTask(taskID, err)
		return
	}

	s.updateTask(taskID, model.BackupTaskRunning, 20, "preparing restore")
	if err := s.ensureDatabaseExists(ctx, profile, req.TargetDatabase); err != nil {
		s.failTask(taskID, err)
		return
	}

	reader, _, err := s.OpenBackupReadCloser(req.Database, req.FileName)
	if err != nil {
		s.failTask(taskID, err)
		return
	}
	defer reader.Close()

	restoreReader := io.Reader(reader)
	if record.Compressed {
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			s.failTask(taskID, err)
			return
		}
		defer gzipReader.Close()
		restoreReader = gzipReader
	}

	s.updateTask(taskID, model.BackupTaskRunning, 55, "restoring backup")
	if err := s.executeRestore(ctx, mysqlBin, profile, req.TargetDatabase, restoreReader); err != nil {
		s.failTask(taskID, err)
		return
	}

	s.finishTask(taskID, record.FileName)
}

func (s *BackupService) executeDump(ctx context.Context, executable string, profile model.OpenConnectionRequest, req model.CreateBackupRequest, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var writer io.Writer = file
	var gzipWriter *gzip.Writer
	if strings.HasSuffix(strings.ToLower(outputPath), ".gz") {
		gzipWriter = gzip.NewWriter(file)
		defer gzipWriter.Close()
		writer = gzipWriter
	}

	args := []string{
		fmt.Sprintf("--host=%s", strings.TrimSpace(profile.Host)),
		fmt.Sprintf("--port=%d", profile.Port),
		fmt.Sprintf("--user=%s", profile.Username),
		"--default-character-set=utf8mb4",
		"--single-transaction",
		"--routines",
		"--events",
		"--triggers",
		"--skip-lock-tables",
	}
	if req.TableName == "" {
		args = append(args, req.Database)
	} else {
		args = append(args, req.Database, req.TableName)
	}

	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+profile.Password)

	var stderr bytes.Buffer
	cmd.Stdout = writer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysqldump failed: %s", strings.TrimSpace(stderr.String()))
	}

	return nil
}

func (s *BackupService) executeRestore(ctx context.Context, executable string, profile model.OpenConnectionRequest, targetDatabase string, reader io.Reader) error {
	args := []string{
		fmt.Sprintf("--host=%s", strings.TrimSpace(profile.Host)),
		fmt.Sprintf("--port=%d", profile.Port),
		fmt.Sprintf("--user=%s", profile.Username),
		"--default-character-set=utf8mb4",
		targetDatabase,
	}

	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+profile.Password)
	cmd.Stdin = reader

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysql restore failed: %s", strings.TrimSpace(stderr.String()))
	}

	return nil
}

func (s *BackupService) ensureDatabaseExists(ctx context.Context, profile model.OpenConnectionRequest, database string) error {
	configs, err := buildCandidateConfigs(model.OpenConnectionRequest{
		Host:     profile.Host,
		Port:     profile.Port,
		Username: profile.Username,
		Password: profile.Password,
		Database: "",
		Params:   profile.Params,
	})
	if err != nil {
		return err
	}

	var db *sql.DB
	for _, cfg := range configs {
		db, err = sql.Open("mysql", cfg.FormatDSN())
		if err != nil {
			continue
		}
		configureDBPool(db)
		if pingErr := db.PingContext(ctx); pingErr == nil {
			break
		}
		_ = db.Close()
		db = nil
	}
	if db == nil {
		return fmt.Errorf("failed to connect mysql for restore")
	}
	defer db.Close()

	quotedDB, err := quoteMySQLIdentifier(database)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+quotedDB+" CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	return formatMySQLError(err)
}

func (s *BackupService) prepareBackupTarget(req model.CreateBackupRequest) (*backupMetadata, string, error) {
	database := strings.TrimSpace(req.Database)
	if database == "" {
		return nil, "", fmt.Errorf("database is required")
	}

	now := time.Now()
	scope := model.BackupScopeDatabase
	displayScope := "full"
	if strings.TrimSpace(req.TableName) != "" {
		scope = model.BackupScopeTable
		displayScope = sanitizeBackupDisplayName(req.TableName)
	}

	displayName := fmt.Sprintf("%s_%s_%s", sanitizeBackupDisplayName(database), displayScope, now.Format("20060102_150405"))
	fileName := displayName + ".sql"
	if req.Compress {
		fileName += ".gz"
	}

	record := &backupMetadata{
		ID:          uuid.NewString(),
		Database:    database,
		TableName:   strings.TrimSpace(req.TableName),
		Scope:       scope,
		FileName:    fileName,
		DisplayName: displayName,
		Compressed:  req.Compress,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	dir, err := s.ensureDatabaseDir(database)
	if err != nil {
		return nil, "", err
	}

	return record, filepath.Join(dir, fileName), nil
}

func (s *BackupService) saveBackupMetadata(database, fileName string, metadata *backupMetadata) (*model.BackupRecord, error) {
	metadata.FileName = fileName
	metadata.UpdatedAt = time.Now()

	payload, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(s.metadataPath(database, fileName), payload, 0644); err != nil {
		return nil, err
	}

	return s.loadBackupRecord(database, fileName)
}

func (s *BackupService) loadBackupRecord(database, fileName string) (*model.BackupRecord, error) {
	if !isSafeBackupFileName(fileName) {
		return nil, fmt.Errorf("invalid backup file name")
	}

	var metadata backupMetadata
	metaPath := s.metadataPath(database, fileName)
	if metaBytes, err := os.ReadFile(metaPath); err == nil {
		if err := json.Unmarshal(metaBytes, &metadata); err != nil {
			return nil, err
		}
	} else {
		metadata = backupMetadata{
			ID:          uuid.NewString(),
			Database:    database,
			FileName:    fileName,
			DisplayName: strings.TrimSuffix(strings.TrimSuffix(fileName, ".gz"), ".sql"),
			Compressed:  strings.HasSuffix(strings.ToLower(fileName), ".gz"),
			Scope:       model.BackupScopeDatabase,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	dir, err := s.ensureDatabaseDir(database)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(filepath.Join(dir, fileName))
	if err != nil {
		return nil, err
	}

	return &model.BackupRecord{
		ID:          metadata.ID,
		Database:    metadata.Database,
		TableName:   metadata.TableName,
		Scope:       metadata.Scope,
		FileName:    fileName,
		DisplayName: metadata.DisplayName,
		Size:        stat.Size(),
		Compressed:  metadata.Compressed,
		CreatedAt:   metadata.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   stat.ModTime().Format(time.RFC3339),
	}, nil
}

func (s *BackupService) ensureDatabaseDir(database string) (string, error) {
	trimmed := strings.TrimSpace(database)
	if trimmed == "" {
		return "", fmt.Errorf("database is required")
	}

	root := filepath.Join(s.baseDir, "storage", "backups", trimmed)
	if err := os.MkdirAll(root, 0755); err != nil {
		return "", err
	}
	return root, nil
}

func (s *BackupService) metadataPath(database, fileName string) string {
	dir, _ := s.ensureDatabaseDir(database)
	return filepath.Join(dir, fileName+".meta.json")
}

func (s *BackupService) newTask(taskType, database, targetDatabase, tableName, fileName string) *model.BackupTask {
	now := time.Now().Format(time.RFC3339)
	task := &model.BackupTask{
		ID:             uuid.NewString(),
		Type:           taskType,
		Database:       database,
		TargetDatabase: targetDatabase,
		TableName:      tableName,
		FileName:       fileName,
		Status:         model.BackupTaskPending,
		Progress:       0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	s.tasksMu.Lock()
	s.tasks[task.ID] = task
	s.tasksMu.Unlock()
	return task
}

func (s *BackupService) updateTask(id string, status model.BackupTaskStatus, progress int, message string) {
	s.tasksMu.Lock()
	defer s.tasksMu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return
	}
	task.Status = status
	task.Progress = progress
	task.Message = message
	task.UpdatedAt = time.Now().Format(time.RFC3339)
}

func (s *BackupService) finishTask(id, fileName string) {
	s.tasksMu.Lock()
	defer s.tasksMu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return
	}
	now := time.Now().Format(time.RFC3339)
	task.Status = model.BackupTaskSuccess
	task.Progress = 100
	task.FileName = fileName
	task.Message = "done"
	task.UpdatedAt = now
	task.CompletedAt = now
}

func (s *BackupService) failTask(id string, err error) {
	s.tasksMu.Lock()
	defer s.tasksMu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return
	}
	now := time.Now().Format(time.RFC3339)
	task.Status = model.BackupTaskFailed
	task.Progress = max(task.Progress, 5)
	task.Message = err.Error()
	task.UpdatedAt = now
	task.CompletedAt = now
}

func resolveMySQLExecutable(base string) (string, error) {
	if path, err := exec.LookPath(base); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("%s executable not found in PATH", base)
}

func sanitizeBackupDisplayName(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	replacer := strings.NewReplacer("\\", "_", "/", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_", " ", "_")
	normalized := replacer.Replace(trimmed)
	return strings.Trim(normalized, "._")
}

func buildRenamedBackupFileName(displayName, oldFileName string) string {
	lower := strings.ToLower(oldFileName)
	switch {
	case strings.HasSuffix(lower, ".sql.gz"):
		return displayName + ".sql.gz"
	case strings.HasSuffix(lower, ".sql"):
		return displayName + ".sql"
	default:
		return displayName
	}
}

func isSafeBackupFileName(fileName string) bool {
	if strings.TrimSpace(fileName) == "" {
		return false
	}
	if strings.Contains(fileName, "..") || strings.ContainsAny(fileName, `/\`) {
		return false
	}
	return true
}

func recordToMetadata(record model.BackupRecord) *backupMetadata {
	createdAt, _ := time.Parse(time.RFC3339, record.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, record.UpdatedAt)
	return &backupMetadata{
		ID:          record.ID,
		Database:    record.Database,
		TableName:   record.TableName,
		Scope:       record.Scope,
		FileName:    record.FileName,
		DisplayName: record.DisplayName,
		Compressed:  record.Compressed,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

