package model

type ExecuteQueryRequest struct {
	SQL      string `json:"sql" binding:"required"`
	Database string `json:"database"`
}

type ExecuteBatchRequest struct {
	Database         string   `json:"database"`
	Statements       []string `json:"statements"`
	InsertStatements []string `json:"insertStatements"`
	UpdateStatements []string `json:"updateStatements"`
	DeleteStatements []string `json:"deleteStatements"`
}

type TableDataResponse struct {
	Columns   []string                 `json:"columns"`
	Rows      []map[string]interface{} `json:"rows"`
	Limit     int                      `json:"limit,omitempty"`
	Offset    int                      `json:"offset,omitempty"`
	Total     int                      `json:"total,omitempty"`
	SortBy    string                   `json:"sortBy,omitempty"`
	SortOrder string                   `json:"sortOrder,omitempty"`
	Keyword   string                   `json:"keyword,omitempty"`
}

type ExecuteResultResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

type ExecuteBatchResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AffectedRows int64  `json:"affected_rows"`
}
