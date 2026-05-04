package model

type SecurityPrincipalKind string

const (
	SecurityPrincipalUser SecurityPrincipalKind = "user"
	SecurityPrincipalRole SecurityPrincipalKind = "role"
)

type SecurityPrincipalRef struct {
	User string `json:"user"`
	Host string `json:"host"`
}

type SecurityScopePrivileges struct {
	Database   string   `json:"database"`
	Table      string   `json:"table,omitempty"`
	Column     string   `json:"column,omitempty"`
	Privileges []string `json:"privileges"`
}

type SecurityPrincipalSummary struct {
	User             string                `json:"user"`
	Host             string                `json:"host"`
	Kind             SecurityPrincipalKind `json:"kind"`
	Locked           bool                  `json:"locked"`
	PasswordExpired  bool                  `json:"passwordExpired"`
	Plugin           string                `json:"plugin"`
	PrivilegeSummary string                `json:"privilegeSummary"`
	PrivilegeDetails string                `json:"privilegeDetails"`
}

type SecurityCapabilities struct {
	Version       string `json:"version"`
	SupportsRoles bool   `json:"supportsRoles"`
}

type SecurityOverview struct {
	Capabilities SecurityCapabilities      `json:"capabilities"`
	Users        []SecurityPrincipalSummary `json:"users"`
	Roles        []SecurityPrincipalSummary `json:"roles"`
}

type SecurityPrincipalDetail struct {
	User             string                  `json:"user"`
	Host             string                  `json:"host"`
	Kind             SecurityPrincipalKind   `json:"kind"`
	Locked           bool                    `json:"locked"`
	PasswordExpired  bool                    `json:"passwordExpired"`
	Plugin           string                  `json:"plugin"`
	GlobalPrivileges []string                `json:"globalPrivileges"`
	SchemaPrivileges []SecurityScopePrivileges `json:"schemaPrivileges"`
	TablePrivileges  []SecurityScopePrivileges `json:"tablePrivileges"`
	ColumnPrivileges []SecurityScopePrivileges `json:"columnPrivileges"`
	Roles            []SecurityPrincipalRef  `json:"roles"`
	GrantStatements  []string                `json:"grantStatements"`
}

type GetSecurityPrincipalRequest struct {
	User string `form:"user" binding:"required"`
	Host string `form:"host" binding:"required"`
	Kind string `form:"kind"`
}

type UpsertSecurityPrincipalRequest struct {
	OriginalUser      string                  `json:"originalUser"`
	OriginalHost      string                  `json:"originalHost"`
	User              string                  `json:"user" binding:"required"`
	Host              string                  `json:"host" binding:"required"`
	Kind              SecurityPrincipalKind   `json:"kind"`
	Password          string                  `json:"password"`
	PasswordChanged   bool                    `json:"passwordChanged"`
	Locked            bool                    `json:"locked"`
	PasswordExpired   bool                    `json:"passwordExpired"`
	GlobalPrivileges  []string                `json:"globalPrivileges"`
	SchemaPrivileges  []SecurityScopePrivileges `json:"schemaPrivileges"`
	TablePrivileges   []SecurityScopePrivileges `json:"tablePrivileges"`
	ColumnPrivileges  []SecurityScopePrivileges `json:"columnPrivileges"`
	Roles             []SecurityPrincipalRef  `json:"roles"`
}

type DeleteSecurityPrincipalRequest struct {
	User string                `json:"user" binding:"required"`
	Host string                `json:"host" binding:"required"`
	Kind SecurityPrincipalKind `json:"kind"`
}

type CloneSecurityPrincipalRequest struct {
	SourceUser string                `json:"sourceUser" binding:"required"`
	SourceHost string                `json:"sourceHost" binding:"required"`
	TargetUser string                `json:"targetUser" binding:"required"`
	TargetHost string                `json:"targetHost" binding:"required"`
	TargetKind SecurityPrincipalKind `json:"targetKind"`
	Password   string                `json:"password"`
}

type RevokeAllSecurityPrincipalRequest struct {
	User string                `json:"user" binding:"required"`
	Host string                `json:"host" binding:"required"`
	Kind SecurityPrincipalKind `json:"kind"`
}
