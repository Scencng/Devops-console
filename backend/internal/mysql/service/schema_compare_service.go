package service

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"devops-console-backend/internal/mysql/model"
)

type SchemaCompareService struct{}

func NewSchemaCompareService() *SchemaCompareService {
	return &SchemaCompareService{}
}

type tableSnapshot struct {
	Database    string
	Table       string
	CreateSQL   string
	Options     tableOptions
	Columns     []columnSnapshot
	columnMap   map[string]columnSnapshot
	Indexes     map[string]indexSnapshot
	ForeignKeys map[string]foreignKeySnapshot
}

type tableOptions struct {
	Engine    string
	Collation string
	Charset   string
}

type columnSnapshot struct {
	Name         string
	Position     int
	ColumnType   string
	Nullable     bool
	DefaultValue sql.NullString
	Extra        string
	Comment      string
	Charset      sql.NullString
	Collation    sql.NullString
	Signature    string
}

type indexSnapshot struct {
	Name      string
	NonUnique bool
	IndexType string
	Columns   []indexColumnSnapshot
	Signature string
}

type indexColumnSnapshot struct {
	Name    string
	SubPart sql.NullInt64
}

type foreignKeySnapshot struct {
	Name       string
	Columns    []string
	RefSchema  string
	RefTable   string
	RefColumns []string
	UpdateRule string
	DeleteRule string
	Signature  string
}

func (s *SchemaCompareService) Compare(ctx context.Context, db *sql.DB, req model.SchemaCompareRequest) (*model.SchemaCompareResponse, error) {
	switch req.Scope {
	case model.SchemaCompareDatabase:
		return s.compareDatabase(ctx, db, req)
	case model.SchemaCompareTable:
		return s.compareTable(ctx, db, req)
	default:
		return nil, fmt.Errorf("unsupported compare scope: %s", req.Scope)
	}
}

func (s *SchemaCompareService) compareDatabase(ctx context.Context, db *sql.DB, req model.SchemaCompareRequest) (*model.SchemaCompareResponse, error) {
	sourceTables, err := listBaseTables(ctx, db, req.SourceDatabase)
	if err != nil {
		return nil, err
	}

	targetTables, err := listBaseTables(ctx, db, req.TargetDatabase)
	if err != nil {
		return nil, err
	}

	sourceSet := make(map[string]struct{}, len(sourceTables))
	targetSet := make(map[string]struct{}, len(targetTables))
	for _, name := range sourceTables {
		sourceSet[name] = struct{}{}
	}
	for _, name := range targetTables {
		targetSet[name] = struct{}{}
	}

	allTables := append([]string{}, sourceTables...)
	for _, name := range targetTables {
		if _, ok := sourceSet[name]; !ok {
			allTables = append(allTables, name)
		}
	}
	slices.Sort(allTables)

	items := make([]model.SchemaDiffItem, 0)
	for _, tableName := range allTables {
		_, inSource := sourceSet[tableName]
		_, inTarget := targetSet[tableName]
		switch {
		case inSource && !inTarget:
			sourceSnapshot, err := loadTableSnapshot(ctx, db, req.SourceDatabase, tableName)
			if err != nil {
				return nil, err
			}
			items = append(items, buildCreateTableItem(*sourceSnapshot, req.TargetDatabase, tableName))
		case !inSource && inTarget:
			items = append(items, buildDropTableItem(req.TargetDatabase, tableName))
		case inSource && inTarget:
			sourceSnapshot, err := loadTableSnapshot(ctx, db, req.SourceDatabase, tableName)
			if err != nil {
				return nil, err
			}
			targetSnapshot, err := loadTableSnapshot(ctx, db, req.TargetDatabase, tableName)
			if err != nil {
				return nil, err
			}
			items = append(items, compareTableSnapshots(*sourceSnapshot, *targetSnapshot, req.TargetDatabase)...)
		}
	}

	return &model.SchemaCompareResponse{
		Scope:          req.Scope,
		SourceDatabase: req.SourceDatabase,
		TargetDatabase: req.TargetDatabase,
		Items:          items,
	}, nil
}

func (s *SchemaCompareService) compareTable(ctx context.Context, db *sql.DB, req model.SchemaCompareRequest) (*model.SchemaCompareResponse, error) {
	if strings.TrimSpace(req.SourceTable) == "" || strings.TrimSpace(req.TargetTable) == "" {
		return nil, fmt.Errorf("sourceTable and targetTable are required for table compare")
	}

	sourceSnapshot, err := loadTableSnapshot(ctx, db, req.SourceDatabase, req.SourceTable)
	if err != nil {
		return nil, err
	}
	targetSnapshot, err := loadTableSnapshot(ctx, db, req.TargetDatabase, req.TargetTable)
	if err != nil {
		return nil, err
	}

	items := compareTableSnapshots(*sourceSnapshot, *targetSnapshot, req.TargetDatabase)
	return &model.SchemaCompareResponse{
		Scope:          req.Scope,
		SourceDatabase: req.SourceDatabase,
		SourceTable:    req.SourceTable,
		TargetDatabase: req.TargetDatabase,
		TargetTable:    req.TargetTable,
		Items:          items,
	}, nil
}

func listBaseTables(ctx context.Context, db *sql.DB, database string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT TABLE_NAME
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		  AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`, database)
	if err != nil {
		return nil, formatMySQLError(err)
	}
	defer rows.Close()

	result := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, formatMySQLError(err)
		}
		result = append(result, name)
	}

	if err := rows.Err(); err != nil {
		return nil, formatMySQLError(err)
	}

	return result, nil
}

func loadTableSnapshot(ctx context.Context, db *sql.DB, database, table string) (*tableSnapshot, error) {
	snapshot := &tableSnapshot{
		Database:    database,
		Table:       table,
		columnMap:   make(map[string]columnSnapshot),
		Indexes:     make(map[string]indexSnapshot),
		ForeignKeys: make(map[string]foreignKeySnapshot),
	}

	if err := db.QueryRowContext(ctx, fmt.Sprintf("SHOW CREATE TABLE %s.%s", mustQuote(database), mustQuote(table))).Scan(&snapshot.Table, &snapshot.CreateSQL); err != nil {
		return nil, formatMySQLError(err)
	}

	if err := loadTableOptions(ctx, db, snapshot); err != nil {
		return nil, err
	}
	if err := loadColumns(ctx, db, snapshot); err != nil {
		return nil, err
	}
	if err := loadIndexes(ctx, db, snapshot); err != nil {
		return nil, err
	}
	if err := loadForeignKeys(ctx, db, snapshot); err != nil {
		return nil, err
	}

	return snapshot, nil
}

func loadTableOptions(ctx context.Context, db *sql.DB, snapshot *tableSnapshot) error {
	var options tableOptions
	if err := db.QueryRowContext(ctx, `
		SELECT ENGINE, TABLE_COLLATION
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
	`, snapshot.Database, snapshot.Table).Scan(&options.Engine, &options.Collation); err != nil {
		return formatMySQLError(err)
	}
	if options.Collation != "" {
		if index := strings.Index(options.Collation, "_"); index > 0 {
			options.Charset = options.Collation[:index]
		}
	}
	snapshot.Options = options
	return nil
}

func loadColumns(ctx context.Context, db *sql.DB, snapshot *tableSnapshot) error {
	rows, err := db.QueryContext(ctx, `
		SELECT ORDINAL_POSITION, COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, EXTRA, COLUMN_COMMENT, CHARACTER_SET_NAME, COLLATION_NAME
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`, snapshot.Database, snapshot.Table)
	if err != nil {
		return formatMySQLError(err)
	}
	defer rows.Close()

	columns := make([]columnSnapshot, 0)
	for rows.Next() {
		var item columnSnapshot
		var isNullable string
		if err := rows.Scan(
			&item.Position,
			&item.Name,
			&item.ColumnType,
			&isNullable,
			&item.DefaultValue,
			&item.Extra,
			&item.Comment,
			&item.Charset,
			&item.Collation,
		); err != nil {
			return formatMySQLError(err)
		}
		item.Nullable = strings.EqualFold(isNullable, "YES")
		item.Signature = buildColumnSignature(item)
		columns = append(columns, item)
		snapshot.columnMap[item.Name] = item
	}

	if err := rows.Err(); err != nil {
		return formatMySQLError(err)
	}

	snapshot.Columns = columns
	return nil
}

func loadIndexes(ctx context.Context, db *sql.DB, snapshot *tableSnapshot) error {
	rows, err := db.QueryContext(ctx, `
		SELECT INDEX_NAME, NON_UNIQUE, INDEX_TYPE, COLUMN_NAME, SUB_PART
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX
	`, snapshot.Database, snapshot.Table)
	if err != nil {
		return formatMySQLError(err)
	}
	defer rows.Close()

	indexes := make(map[string]indexSnapshot)
	for rows.Next() {
		var indexName, indexType, columnName string
		var nonUnique int
		var subPart sql.NullInt64
		if err := rows.Scan(&indexName, &nonUnique, &indexType, &columnName, &subPart); err != nil {
			return formatMySQLError(err)
		}
		current := indexes[indexName]
		current.Name = indexName
		current.NonUnique = nonUnique == 1
		current.IndexType = indexType
		current.Columns = append(current.Columns, indexColumnSnapshot{Name: columnName, SubPart: subPart})
		indexes[indexName] = current
	}
	if err := rows.Err(); err != nil {
		return formatMySQLError(err)
	}

	for key, item := range indexes {
		item.Signature = buildIndexSignature(item)
		indexes[key] = item
	}
	snapshot.Indexes = indexes
	return nil
}

func loadForeignKeys(ctx context.Context, db *sql.DB, snapshot *tableSnapshot) error {
	rows, err := db.QueryContext(ctx, `
		SELECT
			rc.CONSTRAINT_NAME,
			kcu.COLUMN_NAME,
			kcu.REFERENCED_TABLE_SCHEMA,
			kcu.REFERENCED_TABLE_NAME,
			kcu.REFERENCED_COLUMN_NAME,
			rc.UPDATE_RULE,
			rc.DELETE_RULE
		FROM information_schema.REFERENTIAL_CONSTRAINTS rc
		JOIN information_schema.KEY_COLUMN_USAGE kcu
		  ON rc.CONSTRAINT_SCHEMA = kcu.CONSTRAINT_SCHEMA
		 AND rc.TABLE_NAME = kcu.TABLE_NAME
		 AND rc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME
		WHERE rc.CONSTRAINT_SCHEMA = ? AND rc.TABLE_NAME = ?
		ORDER BY rc.CONSTRAINT_NAME, kcu.ORDINAL_POSITION
	`, snapshot.Database, snapshot.Table)
	if err != nil {
		return formatMySQLError(err)
	}
	defer rows.Close()

	result := make(map[string]foreignKeySnapshot)
	for rows.Next() {
		var name, columnName, refSchema, refTable, refColumn, updateRule, deleteRule string
		if err := rows.Scan(&name, &columnName, &refSchema, &refTable, &refColumn, &updateRule, &deleteRule); err != nil {
			return formatMySQLError(err)
		}
		current := result[name]
		current.Name = name
		current.RefSchema = refSchema
		current.RefTable = refTable
		current.UpdateRule = updateRule
		current.DeleteRule = deleteRule
		current.Columns = append(current.Columns, columnName)
		current.RefColumns = append(current.RefColumns, refColumn)
		result[name] = current
	}
	if err := rows.Err(); err != nil {
		return formatMySQLError(err)
	}

	for key, item := range result {
		item.Signature = buildForeignKeySignature(item)
		result[key] = item
	}
	snapshot.ForeignKeys = result
	return nil
}

func compareTableSnapshots(source, target tableSnapshot, targetDatabase string) []model.SchemaDiffItem {
	items := make([]model.SchemaDiffItem, 0)
	items = append(items, compareTableOptions(source, target, targetDatabase)...)
	items = append(items, compareColumns(source, target, targetDatabase)...)
	items = append(items, compareIndexes(source, target, targetDatabase)...)
	items = append(items, compareForeignKeys(source, target, targetDatabase)...)
	return items
}

func compareTableOptions(source, target tableSnapshot, targetDatabase string) []model.SchemaDiffItem {
	if source.Options.Engine == target.Options.Engine && source.Options.Collation == target.Options.Collation {
		return nil
	}

	statement := fmt.Sprintf(
		"ALTER TABLE %s.%s ENGINE=%s DEFAULT CHARACTER SET %s COLLATE %s",
		mustQuote(targetDatabase),
		mustQuote(target.Table),
		source.Options.Engine,
		source.Options.Charset,
		source.Options.Collation,
	)

	return []model.SchemaDiffItem{{
		ID:          fmt.Sprintf("table-options:%s", target.Table),
		Category:    "table",
		ObjectName:  target.Table,
		Title:       fmt.Sprintf("Table options differ for %s", target.Table),
		Detail:      "Engine or collation is different",
		Status:      model.SchemaDiffModify,
		SourceValue: fmt.Sprintf("ENGINE=%s COLLATE=%s", source.Options.Engine, source.Options.Collation),
		TargetValue: fmt.Sprintf("ENGINE=%s COLLATE=%s", target.Options.Engine, target.Options.Collation),
		Statements:  []string{statement},
		Checked:     true,
		Safe:        true,
	}}
}

func compareColumns(source, target tableSnapshot, targetDatabase string) []model.SchemaDiffItem {
	items := make([]model.SchemaDiffItem, 0)
	targetPositions := make(map[string]int, len(target.Columns))
	for index, column := range target.Columns {
		targetPositions[column.Name] = index
	}

	for index, sourceColumn := range source.Columns {
		targetColumn, exists := target.columnMap[sourceColumn.Name]
		if !exists {
			items = append(items, model.SchemaDiffItem{
				ID:          fmt.Sprintf("column-add:%s:%s", target.Table, sourceColumn.Name),
				Category:    "column",
				ObjectName:  fmt.Sprintf("%s.%s", target.Table, sourceColumn.Name),
				Title:       fmt.Sprintf("Missing column %s", sourceColumn.Name),
				Detail:      fmt.Sprintf("Add column %s to %s", sourceColumn.Name, target.Table),
				Status:      model.SchemaDiffAdd,
				SourceValue: sourceColumn.Signature,
				Statements:  []string{buildAddColumnSQL(targetDatabase, target.Table, sourceColumn, source.Columns, index)},
				Checked:     true,
				Safe:        true,
			})
			continue
		}

		positionDiffers := targetPositions[sourceColumn.Name] != index
		if sourceColumn.Signature != targetColumn.Signature || positionDiffers {
			items = append(items, model.SchemaDiffItem{
				ID:          fmt.Sprintf("column-modify:%s:%s", target.Table, sourceColumn.Name),
				Category:    "column",
				ObjectName:  fmt.Sprintf("%s.%s", target.Table, sourceColumn.Name),
				Title:       fmt.Sprintf("Different column %s", sourceColumn.Name),
				Detail:      fmt.Sprintf("Modify column %s in %s", sourceColumn.Name, target.Table),
				Status:      model.SchemaDiffModify,
				SourceValue: sourceColumn.Signature,
				TargetValue: targetColumn.Signature,
				Statements:  []string{buildModifyColumnSQL(targetDatabase, target.Table, sourceColumn, source.Columns, index)},
				Checked:     true,
				Safe:        true,
			})
		}
	}

	for _, targetColumn := range target.Columns {
		if _, exists := source.columnMap[targetColumn.Name]; exists {
			continue
		}
		items = append(items, model.SchemaDiffItem{
			ID:          fmt.Sprintf("column-drop:%s:%s", target.Table, targetColumn.Name),
			Category:    "column",
			ObjectName:  fmt.Sprintf("%s.%s", target.Table, targetColumn.Name),
			Title:       fmt.Sprintf("Extra column %s", targetColumn.Name),
			Detail:      fmt.Sprintf("Drop column %s from %s", targetColumn.Name, target.Table),
			Status:      model.SchemaDiffRemove,
			TargetValue: targetColumn.Signature,
			Statements: []string{
				fmt.Sprintf("ALTER TABLE %s.%s DROP COLUMN %s", mustQuote(targetDatabase), mustQuote(target.Table), mustQuote(targetColumn.Name)),
			},
			Checked: false,
			Safe:    false,
		})
	}

	return items
}

func compareIndexes(source, target tableSnapshot, targetDatabase string) []model.SchemaDiffItem {
	items := make([]model.SchemaDiffItem, 0)

	for name, sourceIndex := range source.Indexes {
		targetIndex, exists := target.Indexes[name]
		if !exists {
			items = append(items, model.SchemaDiffItem{
				ID:          fmt.Sprintf("index-add:%s:%s", target.Table, name),
				Category:    "index",
				ObjectName:  fmt.Sprintf("%s.%s", target.Table, name),
				Title:       fmt.Sprintf("Missing index %s", name),
				Detail:      fmt.Sprintf("Create index %s on %s", name, target.Table),
				Status:      model.SchemaDiffAdd,
				SourceValue: sourceIndex.Signature,
				Statements:  []string{buildAddIndexSQL(targetDatabase, target.Table, sourceIndex)},
				Checked:     true,
				Safe:        true,
			})
			continue
		}
		if sourceIndex.Signature != targetIndex.Signature {
			items = append(items, model.SchemaDiffItem{
				ID:          fmt.Sprintf("index-modify:%s:%s", target.Table, name),
				Category:    "index",
				ObjectName:  fmt.Sprintf("%s.%s", target.Table, name),
				Title:       fmt.Sprintf("Different index %s", name),
				Detail:      fmt.Sprintf("Rebuild index %s on %s", name, target.Table),
				Status:      model.SchemaDiffModify,
				SourceValue: sourceIndex.Signature,
				TargetValue: targetIndex.Signature,
				Statements:  buildReplaceIndexSQL(targetDatabase, target.Table, sourceIndex, targetIndex),
				Checked:     true,
				Safe:        true,
			})
		}
	}

	for name, targetIndex := range target.Indexes {
		if _, exists := source.Indexes[name]; exists {
			continue
		}
		items = append(items, model.SchemaDiffItem{
			ID:          fmt.Sprintf("index-drop:%s:%s", target.Table, name),
			Category:    "index",
			ObjectName:  fmt.Sprintf("%s.%s", target.Table, name),
			Title:       fmt.Sprintf("Extra index %s", name),
			Detail:      fmt.Sprintf("Drop index %s from %s", name, target.Table),
			Status:      model.SchemaDiffRemove,
			TargetValue: targetIndex.Signature,
			Statements:  []string{buildDropIndexSQL(targetDatabase, target.Table, name)},
			Checked:     false,
			Safe:        false,
		})
	}

	return items
}

func compareForeignKeys(source, target tableSnapshot, targetDatabase string) []model.SchemaDiffItem {
	items := make([]model.SchemaDiffItem, 0)

	for name, sourceFK := range source.ForeignKeys {
		targetFK, exists := target.ForeignKeys[name]
		if !exists {
			items = append(items, model.SchemaDiffItem{
				ID:          fmt.Sprintf("fk-add:%s:%s", target.Table, name),
				Category:    "constraint",
				ObjectName:  fmt.Sprintf("%s.%s", target.Table, name),
				Title:       fmt.Sprintf("Missing foreign key %s", name),
				Detail:      fmt.Sprintf("Create foreign key %s on %s", name, target.Table),
				Status:      model.SchemaDiffAdd,
				SourceValue: sourceFK.Signature,
				Statements:  []string{buildAddForeignKeySQL(targetDatabase, source.Database, target.Table, sourceFK)},
				Checked:     true,
				Safe:        true,
			})
			continue
		}

		if sourceFK.Signature != targetFK.Signature {
			items = append(items, model.SchemaDiffItem{
				ID:          fmt.Sprintf("fk-modify:%s:%s", target.Table, name),
				Category:    "constraint",
				ObjectName:  fmt.Sprintf("%s.%s", target.Table, name),
				Title:       fmt.Sprintf("Different foreign key %s", name),
				Detail:      fmt.Sprintf("Rebuild foreign key %s on %s", name, target.Table),
				Status:      model.SchemaDiffModify,
				SourceValue: sourceFK.Signature,
				TargetValue: targetFK.Signature,
				Statements: []string{
					buildDropForeignKeySQL(targetDatabase, target.Table, name),
					buildAddForeignKeySQL(targetDatabase, source.Database, target.Table, sourceFK),
				},
				Checked: true,
				Safe:    true,
			})
		}
	}

	for name, targetFK := range target.ForeignKeys {
		if _, exists := source.ForeignKeys[name]; exists {
			continue
		}
		items = append(items, model.SchemaDiffItem{
			ID:          fmt.Sprintf("fk-drop:%s:%s", target.Table, name),
			Category:    "constraint",
			ObjectName:  fmt.Sprintf("%s.%s", target.Table, name),
			Title:       fmt.Sprintf("Extra foreign key %s", name),
			Detail:      fmt.Sprintf("Drop foreign key %s from %s", name, target.Table),
			Status:      model.SchemaDiffRemove,
			TargetValue: targetFK.Signature,
			Statements:  []string{buildDropForeignKeySQL(targetDatabase, target.Table, name)},
			Checked:     false,
			Safe:        false,
		})
	}

	return items
}

func buildCreateTableItem(source tableSnapshot, targetDatabase, targetTable string) model.SchemaDiffItem {
	return model.SchemaDiffItem{
		ID:          fmt.Sprintf("table-create:%s", targetTable),
		Category:    "table",
		ObjectName:  targetTable,
		Title:       fmt.Sprintf("Missing table %s", targetTable),
		Detail:      fmt.Sprintf("Create table %s in target database", targetTable),
		Status:      model.SchemaDiffAdd,
		SourceValue: fmt.Sprintf("%s.%s", source.Database, source.Table),
		Statements:  []string{rewriteCreateTableName(source.CreateSQL, source.Table, targetTable)},
		Checked:     true,
		Safe:        true,
	}
}

func buildDropTableItem(targetDatabase, table string) model.SchemaDiffItem {
	return model.SchemaDiffItem{
		ID:          fmt.Sprintf("table-drop:%s", table),
		Category:    "table",
		ObjectName:  table,
		Title:       fmt.Sprintf("Extra table %s", table),
		Detail:      fmt.Sprintf("Drop table %s from target database", table),
		Status:      model.SchemaDiffRemove,
		TargetValue: fmt.Sprintf("%s.%s", targetDatabase, table),
		Statements:  []string{fmt.Sprintf("DROP TABLE %s.%s", mustQuote(targetDatabase), mustQuote(table))},
		Checked:     false,
		Safe:        false,
	}
}

func buildAddColumnSQL(targetDatabase, targetTable string, column columnSnapshot, sourceColumns []columnSnapshot, index int) string {
	return fmt.Sprintf(
		"ALTER TABLE %s.%s ADD COLUMN %s%s",
		mustQuote(targetDatabase),
		mustQuote(targetTable),
		buildColumnDefinition(column),
		buildColumnPositionClause(sourceColumns, index),
	)
}

func buildModifyColumnSQL(targetDatabase, targetTable string, column columnSnapshot, sourceColumns []columnSnapshot, index int) string {
	return fmt.Sprintf(
		"ALTER TABLE %s.%s MODIFY COLUMN %s%s",
		mustQuote(targetDatabase),
		mustQuote(targetTable),
		buildColumnDefinition(column),
		buildColumnPositionClause(sourceColumns, index),
	)
}

func buildColumnDefinition(column columnSnapshot) string {
	parts := []string{mustQuote(column.Name), column.ColumnType}
	if column.Charset.Valid && column.Charset.String != "" && isTextColumnType(column.ColumnType) {
		parts = append(parts, "CHARACTER SET "+column.Charset.String)
	}
	if column.Collation.Valid && column.Collation.String != "" && isTextColumnType(column.ColumnType) {
		parts = append(parts, "COLLATE "+column.Collation.String)
	}
	if column.Nullable {
		parts = append(parts, "NULL")
	} else {
		parts = append(parts, "NOT NULL")
	}
	if column.DefaultValue.Valid {
		parts = append(parts, "DEFAULT "+formatColumnDefault(column.DefaultValue.String))
	} else if column.Nullable {
		parts = append(parts, "DEFAULT NULL")
	}
	if trimmedExtra := strings.TrimSpace(column.Extra); trimmedExtra != "" {
		parts = append(parts, strings.ToUpper(trimmedExtra))
	}
	if trimmedComment := strings.TrimSpace(column.Comment); trimmedComment != "" {
		parts = append(parts, "COMMENT "+formatSQLString(trimmedComment))
	}
	return strings.Join(parts, " ")
}

func buildColumnPositionClause(sourceColumns []columnSnapshot, index int) string {
	if index == 0 {
		return " FIRST"
	}
	return " AFTER " + mustQuote(sourceColumns[index-1].Name)
}

func buildAddIndexSQL(targetDatabase, targetTable string, index indexSnapshot) string {
	return fmt.Sprintf(
		"ALTER TABLE %s.%s ADD %s",
		mustQuote(targetDatabase),
		mustQuote(targetTable),
		buildIndexDefinition(index),
	)
}

func buildReplaceIndexSQL(targetDatabase, targetTable string, source, target indexSnapshot) []string {
	return []string{
		buildDropIndexSQL(targetDatabase, targetTable, target.Name),
		buildAddIndexSQL(targetDatabase, targetTable, source),
	}
}

func buildDropIndexSQL(targetDatabase, targetTable, name string) string {
	if strings.EqualFold(name, "PRIMARY") {
		return fmt.Sprintf("ALTER TABLE %s.%s DROP PRIMARY KEY", mustQuote(targetDatabase), mustQuote(targetTable))
	}
	return fmt.Sprintf("ALTER TABLE %s.%s DROP INDEX %s", mustQuote(targetDatabase), mustQuote(targetTable), mustQuote(name))
}

func buildIndexDefinition(index indexSnapshot) string {
	columnSQL := make([]string, 0, len(index.Columns))
	for _, column := range index.Columns {
		item := mustQuote(column.Name)
		if column.SubPart.Valid {
			item = fmt.Sprintf("%s(%d)", item, column.SubPart.Int64)
		}
		columnSQL = append(columnSQL, item)
	}
	columnList := strings.Join(columnSQL, ", ")

	switch strings.ToUpper(index.IndexType) {
	case "FULLTEXT":
		return fmt.Sprintf("FULLTEXT KEY %s (%s)", mustQuote(index.Name), columnList)
	case "SPATIAL":
		return fmt.Sprintf("SPATIAL KEY %s (%s)", mustQuote(index.Name), columnList)
	default:
		if strings.EqualFold(index.Name, "PRIMARY") {
			return fmt.Sprintf("PRIMARY KEY (%s)", columnList)
		}
		if !index.NonUnique {
			return fmt.Sprintf("UNIQUE KEY %s (%s)", mustQuote(index.Name), columnList)
		}
		return fmt.Sprintf("KEY %s (%s)", mustQuote(index.Name), columnList)
	}
}

func buildAddForeignKeySQL(targetDatabase, sourceDatabase, targetTable string, fk foreignKeySnapshot) string {
	refSchema := fk.RefSchema
	if strings.EqualFold(refSchema, sourceDatabase) {
		refSchema = targetDatabase
	}
	columnSQL := quoteIdentifiers(fk.Columns)
	refColumnSQL := quoteIdentifiers(fk.RefColumns)
	return fmt.Sprintf(
		"ALTER TABLE %s.%s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s.%s (%s) ON UPDATE %s ON DELETE %s",
		mustQuote(targetDatabase),
		mustQuote(targetTable),
		mustQuote(fk.Name),
		columnSQL,
		mustQuote(refSchema),
		mustQuote(fk.RefTable),
		refColumnSQL,
		fk.UpdateRule,
		fk.DeleteRule,
	)
}

func buildDropForeignKeySQL(targetDatabase, targetTable, name string) string {
	return fmt.Sprintf("ALTER TABLE %s.%s DROP FOREIGN KEY %s", mustQuote(targetDatabase), mustQuote(targetTable), mustQuote(name))
}

func buildColumnSignature(column columnSnapshot) string {
	defaultValue := "<nil>"
	if column.DefaultValue.Valid {
		defaultValue = column.DefaultValue.String
	}
	return strings.Join([]string{
		strings.ToLower(column.Name),
		strings.ToUpper(column.ColumnType),
		fmt.Sprintf("nullable=%t", column.Nullable),
		"default=" + defaultValue,
		"extra=" + strings.ToUpper(strings.TrimSpace(column.Extra)),
		"comment=" + column.Comment,
		"charset=" + column.Charset.String,
		"collation=" + column.Collation.String,
	}, "|")
}

func buildIndexSignature(index indexSnapshot) string {
	columnParts := make([]string, 0, len(index.Columns))
	for _, column := range index.Columns {
		part := column.Name
		if column.SubPart.Valid {
			part = fmt.Sprintf("%s(%d)", part, column.SubPart.Int64)
		}
		columnParts = append(columnParts, part)
	}
	return fmt.Sprintf("%s|%t|%s|%s", strings.ToUpper(index.Name), index.NonUnique, strings.ToUpper(index.IndexType), strings.Join(columnParts, ","))
}

func buildForeignKeySignature(fk foreignKeySnapshot) string {
	return strings.Join([]string{
		fk.Name,
		strings.Join(fk.Columns, ","),
		fk.RefSchema,
		fk.RefTable,
		strings.Join(fk.RefColumns, ","),
		fk.UpdateRule,
		fk.DeleteRule,
	}, "|")
}

func quoteIdentifiers(items []string) string {
	quoted := make([]string, 0, len(items))
	for _, item := range items {
		quoted = append(quoted, mustQuote(item))
	}
	return strings.Join(quoted, ", ")
}

func rewriteCreateTableName(createSQL, sourceTable, targetTable string) string {
	pattern := regexp.MustCompile("(?i)^CREATE TABLE `?" + regexp.QuoteMeta(sourceTable) + "`?")
	return pattern.ReplaceAllString(createSQL, "CREATE TABLE "+mustQuote(targetTable))
}

func formatColumnDefault(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "''"
	}
	upper := strings.ToUpper(normalized)
	switch upper {
	case "CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP()", "NULL", "NOW()":
		return upper
	}
	if regexp.MustCompile(`^-?\d+(\.\d+)?$`).MatchString(normalized) {
		return normalized
	}
	if strings.HasPrefix(normalized, "b'") || strings.HasPrefix(normalized, "B'") {
		return normalized
	}
	return formatSQLString(normalized)
}

func formatSQLString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func mustQuote(identifier string) string {
	quoted, _ := quoteMySQLIdentifier(identifier)
	return quoted
}

func isTextColumnType(columnType string) bool {
	upper := strings.ToUpper(columnType)
	return strings.Contains(upper, "CHAR") || strings.Contains(upper, "TEXT") || strings.Contains(upper, "ENUM") || strings.Contains(upper, "SET")
}

