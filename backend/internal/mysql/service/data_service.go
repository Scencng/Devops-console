package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"devops-console-backend/internal/mysql/model"
)

type DataService struct{}

func NewDataService() *DataService {
	return &DataService{}
}

func (s *DataService) GetTableData(ctx context.Context, db *sql.DB, databaseName, tableName string, limit, offset int, keyword, sortBy, sortOrder string) (*model.TableDataResponse, error) {
	quotedDB, err := quoteMySQLIdentifier(databaseName)
	if err != nil {
		return nil, err
	}

	quotedTable, err := quoteMySQLIdentifier(tableName)
	if err != nil {
		return nil, err
	}

	columns, err := s.listTableColumns(ctx, db, databaseName, tableName)
	if err != nil {
		return nil, err
	}

	whereClause, whereArgs, err := buildKeywordFilter(columns, strings.TrimSpace(keyword))
	if err != nil {
		return nil, err
	}
	orderClause, err := buildOrderClause(columns, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s%s", quotedDB, quotedTable, whereClause)
	var total int
	if err := db.QueryRowContext(ctx, countQuery, whereArgs...).Scan(&total); err != nil {
		return nil, formatMySQLError(err)
	}

	query := fmt.Sprintf("SELECT * FROM %s.%s%s%s LIMIT ? OFFSET ?", quotedDB, quotedTable, whereClause, orderClause)
	queryArgs := append(append([]any{}, whereArgs...), limit, offset)
	rows, err := db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, formatMySQLError(err)
	}
	defer rows.Close()

	result, err := scanRows(rows)
	if err != nil {
		return nil, err
	}

	result.Limit = limit
	result.Offset = offset
	result.Total = total
	result.SortBy = strings.TrimSpace(sortBy)
	result.SortOrder = normalizeSortOrder(sortOrder)
	result.Keyword = strings.TrimSpace(keyword)
	return result, nil
}

func (s *DataService) listTableColumns(ctx context.Context, db *sql.DB, databaseName, tableName string) ([]string, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT COLUMN_NAME
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`,
		databaseName,
		tableName,
	)
	if err != nil {
		return nil, formatMySQLError(err)
	}
	defer rows.Close()

	columns := make([]string, 0)
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return nil, formatMySQLError(err)
		}

		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		return nil, formatMySQLError(err)
	}

	return columns, nil
}

func buildKeywordFilter(columns []string, keyword string) (string, []any, error) {
	trimmed := strings.TrimSpace(keyword)
	if trimmed == "" || len(columns) == 0 {
		return "", nil, nil
	}

	if clause, args, matched, err := buildExpressionFilter(columns, trimmed); matched {
		if err != nil {
			return "", nil, err
		}

		return clause, args, nil
	}

	conditions := make([]string, 0, len(columns))
	args := make([]any, 0, len(columns))
	likeValue := "%" + trimmed + "%"

	for _, column := range columns {
		quotedColumn, err := quoteMySQLIdentifier(column)
		if err != nil {
			continue
		}

		conditions = append(conditions, fmt.Sprintf("CAST(%s AS CHAR) LIKE ?", quotedColumn))
		args = append(args, likeValue)
	}

	if len(conditions) == 0 {
		return "", nil, nil
	}

	return " WHERE " + strings.Join(conditions, " OR "), args, nil
}

func buildExpressionFilter(columns []string, expression string) (string, []any, bool, error) {
	tokens := splitFilterExpression(expression)
	if len(tokens) == 0 {
		return "", nil, false, nil
	}

	availableColumns := make(map[string]string, len(columns))
	for _, column := range columns {
		availableColumns[strings.ToLower(strings.TrimSpace(column))] = column
	}

	clauses := make([]string, 0, len(tokens))
	args := make([]any, 0, len(tokens))
	matchedAny := false

	for index, token := range tokens {
		upperToken := strings.ToUpper(strings.TrimSpace(token))
		if upperToken == "AND" || upperToken == "OR" {
			if index == 0 || index == len(tokens)-1 {
				return "", nil, true, errors.New("invalid filter expression")
			}

			clauses = append(clauses, upperToken)
			continue
		}

		condition, conditionArgs, matched, err := parseFilterCondition(availableColumns, token)
		if !matched {
			return "", nil, false, nil
		}
		if err != nil {
			return "", nil, true, err
		}

		matchedAny = true
		clauses = append(clauses, condition)
		args = append(args, conditionArgs...)
	}

	if !matchedAny {
		return "", nil, false, nil
	}

	return " WHERE " + strings.Join(clauses, " "), args, true, nil
}

func splitFilterExpression(expression string) []string {
	normalized := strings.TrimSpace(expression)
	if normalized == "" {
		return nil
	}

	parts := strings.Fields(normalized)
	if len(parts) < 3 {
		return nil
	}

	tokens := make([]string, 0)
	current := make([]string, 0)
	for _, part := range parts {
		upperPart := strings.ToUpper(part)
		if upperPart == "AND" || upperPart == "OR" {
			if len(current) > 0 {
				tokens = append(tokens, strings.Join(current, " "))
				current = current[:0]
			}
			tokens = append(tokens, upperPart)
			continue
		}

		current = append(current, part)
	}

	if len(current) > 0 {
		tokens = append(tokens, strings.Join(current, " "))
	}

	return tokens
}

func parseFilterCondition(columns map[string]string, raw string) (string, []any, bool, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil, false, nil
	}

	operators := []string{">=", "<=", "!=", "=", ">", "<"}
	for _, operator := range operators {
		position := strings.Index(trimmed, operator)
		if position <= 0 {
			continue
		}

		columnName := strings.TrimSpace(trimmed[:position])
		column, ok := columns[strings.ToLower(columnName)]
		if !ok {
			return "", nil, true, fmt.Errorf("unknown filter column: %s", columnName)
		}

		valueLiteral := strings.TrimSpace(trimmed[position+len(operator):])
		value, err := parseFilterValue(valueLiteral)
		if err != nil {
			return "", nil, true, err
		}

		quotedColumn, err := quoteMySQLIdentifier(column)
		if err != nil {
			return "", nil, true, err
		}

		if value == nil {
			switch operator {
			case "=":
				return quotedColumn + " IS NULL", nil, true, nil
			case "!=":
				return quotedColumn + " IS NOT NULL", nil, true, nil
			default:
				return "", nil, true, errors.New("null only supports = or !=")
			}
		}

		return quotedColumn + " " + operator + " ?", []any{value}, true, nil
	}

	parts := strings.Fields(trimmed)
	if len(parts) < 3 {
		return "", nil, false, nil
	}

	columnName := parts[0]
	operator := strings.ToLower(parts[1])
	valueLiteral := strings.TrimSpace(strings.Join(parts[2:], " "))
	column, ok := columns[strings.ToLower(columnName)]
	if !ok {
		return "", nil, true, fmt.Errorf("unknown filter column: %s", columnName)
	}

	quotedColumn, err := quoteMySQLIdentifier(column)
	if err != nil {
		return "", nil, true, err
	}

	value, err := parseFilterValue(valueLiteral)
	if err != nil {
		return "", nil, true, err
	}

	stringValue := fmt.Sprintf("%v", value)
	switch operator {
	case "like":
		return quotedColumn + " LIKE ?", []any{"%" + stringValue + "%"}, true, nil
	case "contains":
		return quotedColumn + " LIKE ?", []any{"%" + stringValue + "%"}, true, nil
	case "startswith":
		return quotedColumn + " LIKE ?", []any{stringValue + "%"}, true, nil
	case "endswith":
		return quotedColumn + " LIKE ?", []any{"%" + stringValue}, true, nil
	default:
		return "", nil, false, nil
	}
}

func parseFilterValue(raw string) (any, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, errors.New("filter value is required")
	}

	if (strings.HasPrefix(trimmed, "'") && strings.HasSuffix(trimmed, "'")) ||
		(strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"")) {
		return trimmed[1 : len(trimmed)-1], nil
	}

	lower := strings.ToLower(trimmed)
	switch lower {
	case "null":
		return nil, nil
	case "true":
		return true, nil
	case "false":
		return false, nil
	}

	if intValue, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return intValue, nil
	}

	if floatValue, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return floatValue, nil
	}

	return trimmed, nil
}

func buildOrderClause(columns []string, sortBy, sortOrder string) (string, error) {
	trimmedSortBy := strings.TrimSpace(sortBy)
	if trimmedSortBy == "" {
		for _, column := range columns {
			if strings.EqualFold(strings.TrimSpace(column), "id") {
				quotedColumn, err := quoteMySQLIdentifier(column)
				if err != nil {
					return "", err
				}

				return " ORDER BY " + quotedColumn + " ASC", nil
			}
		}

		return "", nil
	}

	for _, column := range columns {
		if column != trimmedSortBy {
			continue
		}

		quotedColumn, err := quoteMySQLIdentifier(trimmedSortBy)
		if err != nil {
			return "", err
		}

		return " ORDER BY " + quotedColumn + " " + strings.ToUpper(normalizeSortOrder(sortOrder)), nil
	}

	return "", fmt.Errorf("unknown sort column: %s", trimmedSortBy)
}

func normalizeSortOrder(sortOrder string) string {
	if strings.EqualFold(strings.TrimSpace(sortOrder), "desc") {
		return "desc"
	}

	return "asc"
}

