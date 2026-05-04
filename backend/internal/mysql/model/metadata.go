package model

type CreateDatabaseRequest struct {
	Name string `json:"name" binding:"required"`
}

type RenameDatabaseRequest struct {
	OldName string `json:"oldName" binding:"required"`
	NewName string `json:"newName" binding:"required"`
}

type DeleteDatabaseRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateTableRequest struct {
	Database string `json:"database" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Columns  string `json:"columns" binding:"required"`
	Options  string `json:"options"`
}

type AutoImportColumn struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required"`
}

type AutoImportTableRequest struct {
	Database string              `json:"database" binding:"required"`
	Name     string              `json:"name" binding:"required"`
	Columns  []AutoImportColumn  `json:"columns" binding:"required"`
	Rows     []map[string]any    `json:"rows"`
}

type AutoImportTableResponse struct {
	Success   bool   `json:"success"`
	TableName string `json:"tableName"`
	RowCount  int    `json:"rowCount"`
}

type RenameTableRequest struct {
	Database string `json:"database" binding:"required"`
	OldName  string `json:"oldName" binding:"required"`
	NewName  string `json:"newName" binding:"required"`
}

type DeleteTableRequest struct {
	Database string `json:"database" binding:"required"`
	Name     string `json:"name" binding:"required"`
}
