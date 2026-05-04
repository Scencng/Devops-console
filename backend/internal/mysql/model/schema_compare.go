package model

type SchemaCompareScope string

const (
	SchemaCompareDatabase SchemaCompareScope = "database"
	SchemaCompareTable    SchemaCompareScope = "table"
)

type SchemaCompareRequest struct {
	Scope          SchemaCompareScope `json:"scope" binding:"required"`
	SourceDatabase string             `json:"sourceDatabase" binding:"required"`
	SourceTable    string             `json:"sourceTable"`
	TargetDatabase string             `json:"targetDatabase" binding:"required"`
	TargetTable    string             `json:"targetTable"`
}

type SchemaCompareResponse struct {
	Scope          SchemaCompareScope `json:"scope"`
	SourceDatabase string             `json:"sourceDatabase"`
	SourceTable    string             `json:"sourceTable,omitempty"`
	TargetDatabase string             `json:"targetDatabase"`
	TargetTable    string             `json:"targetTable,omitempty"`
	Items          []SchemaDiffItem   `json:"items"`
}

type SchemaDiffStatus string

const (
	SchemaDiffAdd    SchemaDiffStatus = "add"
	SchemaDiffRemove SchemaDiffStatus = "remove"
	SchemaDiffModify SchemaDiffStatus = "modify"
)

type SchemaDiffItem struct {
	ID          string           `json:"id"`
	Category    string           `json:"category"`
	ObjectName  string           `json:"objectName"`
	Title       string           `json:"title"`
	Detail      string           `json:"detail"`
	Status      SchemaDiffStatus `json:"status"`
	SourceValue string           `json:"sourceValue,omitempty"`
	TargetValue string           `json:"targetValue,omitempty"`
	Statements  []string         `json:"statements"`
	Checked     bool             `json:"checked"`
	Safe        bool             `json:"safe"`
}
