package service

import (
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"

	"devops-console-backend/internal/mysql/model"
)

var errEmptySQL = errors.New("sql statement is required")

func scanRows(rows *sql.Rows) (*model.TableDataResponse, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, formatMySQLError(err)
	}
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, formatMySQLError(err)
	}

	rowValues := make([]interface{}, len(columns))
	scanTargets := make([]interface{}, len(columns))
	for i := range columns {
		scanTargets[i] = &rowValues[i]
	}

	result := &model.TableDataResponse{
		Columns: columns,
		Rows:    make([]map[string]interface{}, 0),
	}

	for rows.Next() {
		if err := rows.Scan(scanTargets...); err != nil {
			return nil, formatMySQLError(err)
		}

		record := make(map[string]interface{}, len(columns))
		for idx, column := range columns {
			typeName := ""
			if idx < len(columnTypes) && columnTypes[idx] != nil {
				typeName = columnTypes[idx].DatabaseTypeName()
			}
			record[column] = normalizeSQLValue(rowValues[idx], typeName)
		}

		result.Rows = append(result.Rows, record)
	}

	if err := rows.Err(); err != nil {
		return nil, formatMySQLError(err)
	}

	return result, nil
}

func normalizeSQLValue(value interface{}, typeName string) interface{} {
	switch v := value.(type) {
	case nil:
		return nil
	case []byte:
		if isSpatialColumnType(typeName) {
			if text, ok := decodeMySQLPointValue(v); ok {
				return text
			}
			return "鏁版嵁瑙ｆ瀽澶辫触"
		}
		if utf8.Valid(v) {
			return string(v)
		}
		return string(v)
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	default:
		return v
	}
}

func isSpatialColumnType(typeName string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(typeName))
	return strings.Contains(normalized, "POINT") ||
		strings.Contains(normalized, "GEOMETRY")
}

func decodeMySQLPointValue(value []byte) (string, bool) {
	if len(value) == 0 {
		return "", false
	}

	if utf8.Valid(value) {
		text := strings.TrimSpace(string(value))
		if strings.HasPrefix(strings.ToUpper(text), "POINT") {
			return text, true
		}
	}

	switch len(value) {
	case 21:
		return decodeWKBPoint(value)
	case 25:
		return decodeWKBPoint(value[4:])
	default:
		if len(value) > 25 {
			if text, ok := decodeWKBPoint(value[4:]); ok {
				return text, true
			}
		}
		return "", false
	}
}

func decodeWKBPoint(value []byte) (string, bool) {
	if len(value) < 21 {
		return "", false
	}

	var byteOrder binary.ByteOrder
	switch value[0] {
	case 0:
		byteOrder = binary.BigEndian
	case 1:
		byteOrder = binary.LittleEndian
	default:
		return "", false
	}

	geometryType := byteOrder.Uint32(value[1:5])
	if geometryType != 1 {
		return "", false
	}

	x := math.Float64frombits(byteOrder.Uint64(value[5:13]))
	y := math.Float64frombits(byteOrder.Uint64(value[13:21]))
	if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) {
		return "", false
	}

	return fmt.Sprintf("POINT(%s %s)", formatPointCoordinate(x), formatPointCoordinate(y)), true
}

func formatPointCoordinate(value float64) string {
	formatted := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.12f", value), "0"), ".")
	if formatted == "" || formatted == "-0" {
		return "0"
	}
	return formatted
}

func isQueryStatement(statement string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(statement))
	return strings.HasPrefix(normalized, "SELECT") ||
		strings.HasPrefix(normalized, "SHOW") ||
		strings.HasPrefix(normalized, "DESC") ||
		strings.HasPrefix(normalized, "DESCRIBE") ||
		strings.HasPrefix(normalized, "EXPLAIN") ||
		strings.HasPrefix(normalized, "WITH")
}

func formatMySQLError(err error) error {
	if err == nil {
		return nil
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return fmt.Errorf("mysql error %d: %s", mysqlErr.Number, mysqlErr.Message)
	}

	return err
}

func quoteMySQLIdentifier(identifier string) (string, error) {
	trimmed := strings.TrimSpace(identifier)
	if trimmed == "" {
		return "", errors.New("identifier cannot be empty")
	}

	return "`" + strings.ReplaceAll(trimmed, "`", "``") + "`", nil
}

