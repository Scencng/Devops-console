package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"devops-console-backend/internal/mysql/model"
)

type QueryService struct{}

func NewQueryService() *QueryService {
	return &QueryService{}
}

func (s *QueryService) Execute(ctx context.Context, db *sql.DB, req model.ExecuteQueryRequest) (interface{}, error) {
	trimmed := strings.TrimSpace(req.SQL)
	if trimmed == "" {
		return nil, formatMySQLError(errEmptySQL)
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, formatMySQLError(err)
	}
	defer conn.Close()

	if err := useDatabase(ctx, conn, req.Database); err != nil {
		return nil, err
	}

	if isQueryStatement(trimmed) {
		rows, err := conn.QueryContext(ctx, trimmed)
		if err != nil {
			return nil, formatMySQLError(err)
		}
		defer rows.Close()

		return scanRows(rows)
	}

	result, err := conn.ExecContext(ctx, trimmed)
	if err != nil {
		return nil, formatMySQLError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, formatMySQLError(err)
	}

	return model.ExecuteResultResponse{
		RowsAffected: rowsAffected,
	}, nil
}

func (s *QueryService) ExecuteBatch(ctx context.Context, db *sql.DB, statements model.ExecuteBatchRequest) (*model.ExecuteBatchResponse, error) {
	var batch []string
	if len(statements.Statements) > 0 {
		batch = append(batch, statements.Statements...)
	} else {
		batch = append(batch, statements.InsertStatements...)
		batch = append(batch, statements.UpdateStatements...)
		batch = append(batch, statements.DeleteStatements...)
	}

	if len(batch) == 0 {
		return nil, formatMySQLError(errEmptySQL)
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

	if err := useDatabase(ctx, tx, statements.Database); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	var totalAffected int64
	for index, statement := range batch {
		trimmed := strings.TrimSpace(statement)
		if trimmed == "" {
			_ = tx.Rollback()
			return nil, formatMySQLError(fmt.Errorf("sql statement at index %d is required", index))
		}

		result, execErr := tx.ExecContext(ctx, trimmed)
		if execErr != nil {
			_ = tx.Rollback()
			return nil, formatMySQLError(execErr)
		}

		rowsAffected, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			_ = tx.Rollback()
			return nil, formatMySQLError(rowsErr)
		}

		totalAffected += rowsAffected
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, formatMySQLError(err)
	}

	return &model.ExecuteBatchResponse{
		Success:      true,
		Message:      "batch execution succeeded",
		AffectedRows: totalAffected,
	}, nil
}

type sqlExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func useDatabase(ctx context.Context, executor sqlExecutor, database string) error {
	trimmed := strings.TrimSpace(database)
	if trimmed == "" {
		return nil
	}

	quotedDB, err := quoteMySQLIdentifier(trimmed)
	if err != nil {
		return err
	}

	if _, err := executor.ExecContext(ctx, "USE "+quotedDB); err != nil {
		return formatMySQLError(err)
	}

	return nil
}

