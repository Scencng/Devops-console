package service

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"devops-console-backend/internal/mysql/model"
)

type MetadataService struct{}

func NewMetadataService() *MetadataService {
	return &MetadataService{}
}

func (s *MetadataService) ListDatabases(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SHOW DATABASES")
	if err != nil {
		return nil, formatMySQLError(err)
	}
	defer rows.Close()

	systemDatabases := map[string]struct{}{
		"information_schema": {},
		"performance_schema": {},
		"mysql":              {},
		"sys":                {},
	}

	var databases []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, formatMySQLError(err)
		}

		if _, excluded := systemDatabases[name]; excluded {
			continue
		}

		databases = append(databases, name)
	}

	if err := rows.Err(); err != nil {
		return nil, formatMySQLError(err)
	}

	return databases, nil
}

func (s *MetadataService) ListTables(ctx context.Context, db *sql.DB, databaseName string) ([]string, error) {
	quotedDB, err := quoteMySQLIdentifier(databaseName)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SHOW TABLES FROM %s", quotedDB)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, formatMySQLError(err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, formatMySQLError(err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, formatMySQLError(err)
	}

	return tables, nil
}

func (s *MetadataService) CreateDatabase(ctx context.Context, db *sql.DB, databaseName string) error {
	quotedDB, err := quoteMySQLIdentifier(databaseName)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, "CREATE DATABASE "+quotedDB); err != nil {
		return formatMySQLError(err)
	}

	return nil
}

func (s *MetadataService) DeleteDatabase(ctx context.Context, db *sql.DB, databaseName string) error {
	quotedDB, err := quoteMySQLIdentifier(databaseName)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, "DROP DATABASE "+quotedDB); err != nil {
		return formatMySQLError(err)
	}

	return nil
}

func (s *MetadataService) RenameDatabase(ctx context.Context, db *sql.DB, oldName, newName string) error {
	quotedOldDB, err := quoteMySQLIdentifier(oldName)
	if err != nil {
		return err
	}

	quotedNewDB, err := quoteMySQLIdentifier(newName)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, "CREATE DATABASE "+quotedNewDB); err != nil {
		return formatMySQLError(err)
	}

	tables, err := s.ListTables(ctx, db, oldName)
	if err != nil {
		return err
	}

	for _, tableName := range tables {
		quotedTable, quoteErr := quoteMySQLIdentifier(tableName)
		if quoteErr != nil {
			return quoteErr
		}

		renameSQL := fmt.Sprintf(
			"RENAME TABLE %s.%s TO %s.%s",
			quotedOldDB,
			quotedTable,
			quotedNewDB,
			quotedTable,
		)
		if _, execErr := db.ExecContext(ctx, renameSQL); execErr != nil {
			return formatMySQLError(execErr)
		}
	}

	if _, err := db.ExecContext(ctx, "DROP DATABASE "+quotedOldDB); err != nil {
		return formatMySQLError(err)
	}

	return nil
}

func (s *MetadataService) CreateTable(ctx context.Context, db *sql.DB, databaseName, tableName, columnsSQL, optionsSQL string) error {
	quotedDB, err := quoteMySQLIdentifier(databaseName)
	if err != nil {
		return err
	}

	quotedTable, err := quoteMySQLIdentifier(tableName)
	if err != nil {
		return err
	}

	trimmedColumns := strings.TrimSpace(columnsSQL)
	if trimmedColumns == "" {
		return fmt.Errorf("table columns definition is required")
	}

	query := fmt.Sprintf("CREATE TABLE %s.%s (%s)", quotedDB, quotedTable, trimmedColumns)
	if trimmedOptions := strings.TrimSpace(optionsSQL); trimmedOptions != "" {
		query += " " + trimmedOptions
	}
	if _, err := db.ExecContext(ctx, query); err != nil {
		return formatMySQLError(err)
	}

	return nil
}

func (s *MetadataService) AutoImportTable(ctx context.Context, db *sql.DB, req model.AutoImportTableRequest) (*model.AutoImportTableResponse, error) {
	if strings.TrimSpace(req.Database) == "" {
		return nil, fmt.Errorf("database is required")
	}

	baseTableName := strings.TrimSpace(req.Name)
	if baseTableName == "" {
		return nil, fmt.Errorf("table name is required")
	}

	if len(req.Columns) == 0 {
		return nil, fmt.Errorf("at least one import column is required")
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, formatMySQLError(err)
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, formatMySQLError(err)
	}

	if err := useDatabase(ctx, tx, req.Database); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	tableName, err := s.nextAvailableTableName(ctx, tx, req.Database, baseTableName)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	createSQL, orderedColumns, err := buildAutoImportCreateTableSQL(req.Database, tableName, req.Columns)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, createSQL); err != nil {
		_ = tx.Rollback()
		return nil, formatMySQLError(err)
	}

	if len(req.Rows) > 0 {
		insertSQL, err := buildAutoImportInsertSQL(req.Database, tableName, orderedColumns)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		for index, row := range req.Rows {
			args := make([]any, 0, len(orderedColumns))
			for _, column := range orderedColumns {
				normalizedValue, err := normalizeAutoImportValue(column.Type, row[column.Name])
				if err != nil {
					_ = tx.Rollback()
					return nil, fmt.Errorf("import row %d column %s: %w", index+1, column.Name, err)
				}
				args = append(args, normalizedValue)
			}

			if _, err := tx.ExecContext(ctx, insertSQL, args...); err != nil {
				_ = tx.Rollback()
				return nil, formatMySQLError(fmt.Errorf("import row %d failed: %w", index+1, err))
			}
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, formatMySQLError(err)
	}

	return &model.AutoImportTableResponse{
		Success:   true,
		TableName: tableName,
		RowCount:  len(req.Rows),
	}, nil
}

func (s *MetadataService) RenameTable(ctx context.Context, db *sql.DB, databaseName, oldName, newName string) error {
	quotedDB, err := quoteMySQLIdentifier(databaseName)
	if err != nil {
		return err
	}

	quotedOldTable, err := quoteMySQLIdentifier(oldName)
	if err != nil {
		return err
	}

	quotedNewTable, err := quoteMySQLIdentifier(newName)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(
		"RENAME TABLE %s.%s TO %s.%s",
		quotedDB,
		quotedOldTable,
		quotedDB,
		quotedNewTable,
	)
	if _, err := db.ExecContext(ctx, query); err != nil {
		return formatMySQLError(err)
	}

	return nil
}

func (s *MetadataService) DeleteTable(ctx context.Context, db *sql.DB, databaseName, tableName string) error {
	quotedDB, err := quoteMySQLIdentifier(databaseName)
	if err != nil {
		return err
	}

	quotedTable, err := quoteMySQLIdentifier(tableName)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("DROP TABLE %s.%s", quotedDB, quotedTable)
	if _, err := db.ExecContext(ctx, query); err != nil {
		return formatMySQLError(err)
	}

	return nil
}

var autoImportTypePattern = regexp.MustCompile(`^[A-Z]+(?:\(\d+(?:,\d+)?\))?$`)
var autoImportCJKSpacePattern = regexp.MustCompile(`([\p{Han}])\s+([\p{Han}])`)

func buildAutoImportCreateTableSQL(databaseName, tableName string, columns []model.AutoImportColumn) (string, []model.AutoImportColumn, error) {
	quotedDB, err := quoteMySQLIdentifier(databaseName)
	if err != nil {
		return "", nil, err
	}

	quotedTable, err := quoteMySQLIdentifier(tableName)
	if err != nil {
		return "", nil, err
	}

	orderedColumns := make([]model.AutoImportColumn, 0, len(columns))
	seen := make(map[string]struct{}, len(columns))
	definitions := make([]string, 0, len(columns)+2)
	hasPrimaryID := false

	for _, column := range columns {
		name := strings.TrimSpace(column.Name)
		if name == "" {
			return "", nil, fmt.Errorf("import column name is required")
		}

		normalizedName := strings.ToLower(name)
		if _, exists := seen[normalizedName]; exists {
			return "", nil, fmt.Errorf("duplicate import column: %s", name)
		}
		seen[normalizedName] = struct{}{}

		columnType := strings.ToUpper(strings.TrimSpace(column.Type))
		if !autoImportTypePattern.MatchString(columnType) {
			return "", nil, fmt.Errorf("unsupported import column type: %s", column.Type)
		}

		quotedColumn, err := quoteMySQLIdentifier(name)
		if err != nil {
			return "", nil, err
		}

		if normalizedName == "id" {
			hasPrimaryID = true
			primaryKeyType := normalizePrimaryKeyType(columnType)
			if supportsPrimaryKeyAutoIncrement(primaryKeyType) {
				definitions = append(definitions, fmt.Sprintf("%s %s NOT NULL AUTO_INCREMENT", quotedColumn, primaryKeyType))
			} else {
				definitions = append(definitions, fmt.Sprintf("%s %s NOT NULL", quotedColumn, primaryKeyType))
			}
			orderedColumns = append(orderedColumns, model.AutoImportColumn{Name: name, Type: primaryKeyType})
			continue
		}

		definitions = append(definitions, fmt.Sprintf("%s %s NULL", quotedColumn, columnType))
		orderedColumns = append(orderedColumns, model.AutoImportColumn{Name: name, Type: columnType})
	}

	if hasPrimaryID {
		definitions = append(definitions, "PRIMARY KEY (`id`)")
	} else {
		definitions = append([]string{"`id` BIGINT NOT NULL AUTO_INCREMENT"}, definitions...)
		definitions = append(definitions, "PRIMARY KEY (`id`)")
	}

	query := fmt.Sprintf(
		"CREATE TABLE %s.%s (%s) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci",
		quotedDB,
		quotedTable,
		strings.Join(definitions, ", "),
	)

	return query, orderedColumns, nil
}

func normalizePrimaryKeyType(columnType string) string {
	normalized := strings.ToUpper(strings.TrimSpace(columnType))
	switch normalized {
	case "INT", "INTEGER", "TINYINT", "SMALLINT", "MEDIUMINT":
		return "BIGINT"
	default:
		return normalized
	}
}

func supportsPrimaryKeyAutoIncrement(columnType string) bool {
	switch strings.ToUpper(strings.TrimSpace(columnType)) {
	case "BIGINT", "INT", "INTEGER", "TINYINT", "SMALLINT", "MEDIUMINT":
		return true
	default:
		return false
	}
}

func normalizeAutoImportValue(columnType string, value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	normalizedType := strings.ToUpper(strings.TrimSpace(columnType))
	text := normalizeAutoImportTextValue(fmt.Sprintf("%v", value))
	if text == "" {
		return nil, nil
	}

	switch normalizedType {
	case "DATETIME", "TIMESTAMP":
		if parsed, ok := parseAutoImportTime(text); ok {
			return parsed.Format("2006-01-02 15:04:05"), nil
		}
		return nil, fmt.Errorf("invalid datetime value %q", text)
	case "DATE":
		if parsed, ok := parseAutoImportTime(text); ok {
			return parsed.Format("2006-01-02"), nil
		}
		return nil, fmt.Errorf("invalid date value %q", text)
	}

	return text, nil
}

func normalizeAutoImportTextValue(value string) string {
	normalized := strings.ReplaceAll(value, "\u00A0", " ")
	normalized = strings.ReplaceAll(normalized, "\u3000", " ")
	normalized = strings.NewReplacer("\r", " ", "\n", " ", "\t", " ").Replace(normalized)
	normalized = strings.Join(strings.Fields(strings.TrimSpace(normalized)), " ")
	if normalized == "" {
		return ""
	}

	for {
		next := autoImportCJKSpacePattern.ReplaceAllString(normalized, "$1$2")
		if next == normalized {
			return next
		}
		normalized = next
	}
}

func parseAutoImportTime(text string) (time.Time, bool) {
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, text); err == nil {
			return parsed, true
		}
	}

	return time.Time{}, false
}

func buildAutoImportInsertSQL(databaseName, tableName string, columns []model.AutoImportColumn) (string, error) {
	quotedDB, err := quoteMySQLIdentifier(databaseName)
	if err != nil {
		return "", err
	}

	quotedTable, err := quoteMySQLIdentifier(tableName)
	if err != nil {
		return "", err
	}

	quotedColumns := make([]string, 0, len(columns))
	placeholders := make([]string, 0, len(columns))
	for _, column := range columns {
		quotedColumn, err := quoteMySQLIdentifier(column.Name)
		if err != nil {
			return "", err
		}
		quotedColumns = append(quotedColumns, quotedColumn)
		placeholders = append(placeholders, "?")
	}

	return fmt.Sprintf(
		"INSERT INTO %s.%s (%s) VALUES (%s)",
		quotedDB,
		quotedTable,
		strings.Join(quotedColumns, ", "),
		strings.Join(placeholders, ", "),
	), nil
}

func (s *MetadataService) nextAvailableTableName(ctx context.Context, tx *sql.Tx, databaseName, baseTableName string) (string, error) {
	base := strings.TrimSpace(baseTableName)
	if base == "" {
		return "", fmt.Errorf("table name is required")
	}

	candidate := base
	for suffix := 0; suffix < 1000; suffix++ {
		var exists int
		if err := tx.QueryRowContext(
			ctx,
			`SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`,
			databaseName,
			candidate,
		).Scan(&exists); err != nil {
			return "", formatMySQLError(err)
		}

		if exists == 0 {
			return candidate, nil
		}

		candidate = fmt.Sprintf("%s_%d", base, suffix+1)
	}

	return "", fmt.Errorf("unable to allocate a unique table name for %s", base)
}

