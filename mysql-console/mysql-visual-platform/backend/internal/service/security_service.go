package service

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"mydeploy-project/internal/model"
)

type SecurityService struct{}

func NewSecurityService() *SecurityService {
	return &SecurityService{}
}

type mysqlUserFeatures struct {
	hasPlugin          bool
	hasAccountLocked   bool
	hasPasswordExpired bool
	hasIsRole          bool
}

func (s *SecurityService) GetOverview(ctx context.Context, db *sql.DB) (model.SecurityOverview, error) {
	capabilities, features, err := s.getCapabilities(ctx, db)
	if err != nil {
		return model.SecurityOverview{}, err
	}

	users, err := s.listPrincipals(ctx, db, features, model.SecurityPrincipalUser)
	if err != nil {
		return model.SecurityOverview{}, err
	}

	roles := make([]model.SecurityPrincipalSummary, 0)
	if capabilities.SupportsRoles {
		roles, err = s.listPrincipals(ctx, db, features, model.SecurityPrincipalRole)
		if err != nil {
			return model.SecurityOverview{}, err
		}
	}

	return model.SecurityOverview{
		Capabilities: capabilities,
		Users:        users,
		Roles:        roles,
	}, nil
}

func (s *SecurityService) GetPrincipalDetail(ctx context.Context, db *sql.DB, user, host string, kind model.SecurityPrincipalKind) (model.SecurityPrincipalDetail, error) {
	capabilities, features, err := s.getCapabilities(ctx, db)
	if err != nil {
		return model.SecurityPrincipalDetail{}, err
	}

	detail, err := s.loadPrincipalBase(ctx, db, features, user, host, kind)
	if err != nil {
		return model.SecurityPrincipalDetail{}, err
	}

	grantee := formatGrantee(user, host)

	detail.GlobalPrivileges, err = s.listGlobalPrivileges(ctx, db, grantee)
	if err != nil {
		return model.SecurityPrincipalDetail{}, err
	}

	detail.SchemaPrivileges, err = s.listSchemaPrivileges(ctx, db, grantee)
	if err != nil {
		return model.SecurityPrincipalDetail{}, err
	}

	detail.TablePrivileges, err = s.listTablePrivileges(ctx, db, grantee)
	if err != nil {
		return model.SecurityPrincipalDetail{}, err
	}

	detail.ColumnPrivileges, err = s.listColumnPrivileges(ctx, db, grantee)
	if err != nil {
		return model.SecurityPrincipalDetail{}, err
	}

	if capabilities.SupportsRoles {
		detail.Roles, _ = s.listGrantedRoles(ctx, db, user, host)
	}

	detail.GrantStatements, err = s.showGrantStatements(ctx, db, user, host)
	if err != nil {
		return model.SecurityPrincipalDetail{}, err
	}

	return detail, nil
}

func (s *SecurityService) CreatePrincipal(ctx context.Context, db *sql.DB, req model.UpsertSecurityPrincipalRequest) error {
	capabilities, _, err := s.getCapabilities(ctx, db)
	if err != nil {
		return err
	}

	req.Kind = normalizePrincipalKind(req.Kind)
	if req.Kind == model.SecurityPrincipalRole {
		if !capabilities.SupportsRoles {
			return errors.New("current mysql version does not support roles")
		}
		if err := execSecuritySQL(ctx, db, fmt.Sprintf("CREATE ROLE %s", formatGrantee(req.User, req.Host))); err != nil {
			if isGrantTableCompatError(err) {
				return errors.New("current mysql instance cannot create roles because grant tables are in compatibility mode")
			}
			return err
		}
	} else {
		if strings.TrimSpace(req.Password) == "" {
			return errors.New("password is required when creating a user")
		}
		if err := execSecuritySQL(ctx, db, fmt.Sprintf("CREATE USER %s IDENTIFIED BY %s", formatGrantee(req.User, req.Host), quoteString(req.Password))); err != nil {
			if compatErr := s.createPrincipalCompat(ctx, db, req, err); compatErr != nil {
				return compatErr
			}
		}
	}

	if err := s.applyPrincipalSecurity(ctx, db, capabilities, req, false); err != nil {
		return err
	}

	s.writeAuditLog(ctx, db, "create-principal", req.User, req.Host, req.Kind)
	return nil
}

func (s *SecurityService) UpdatePrincipal(ctx context.Context, db *sql.DB, req model.UpsertSecurityPrincipalRequest) error {
	capabilities, _, err := s.getCapabilities(ctx, db)
	if err != nil {
		return err
	}

	req.Kind = normalizePrincipalKind(req.Kind)
	originalUser := strings.TrimSpace(req.OriginalUser)
	originalHost := strings.TrimSpace(req.OriginalHost)
	if originalUser == "" {
		originalUser = req.User
	}
	if originalHost == "" {
		originalHost = req.Host
	}

	if originalUser != req.User || originalHost != req.Host {
		if err := execSecuritySQL(
			ctx,
			db,
			fmt.Sprintf("RENAME USER %s TO %s", formatGrantee(originalUser, originalHost), formatGrantee(req.User, req.Host)),
		); err != nil {
			if compatErr := s.renamePrincipalCompat(ctx, db, req.Kind, originalUser, originalHost, req.User, req.Host, err); compatErr != nil {
				return compatErr
			}
		}
	}

	if err := s.applyPrincipalSecurity(ctx, db, capabilities, req, true); err != nil {
		return err
	}

	s.writeAuditLog(ctx, db, "update-principal", req.User, req.Host, req.Kind)
	return nil
}

func (s *SecurityService) DeletePrincipal(ctx context.Context, db *sql.DB, req model.DeleteSecurityPrincipalRequest) error {
	currentActor, _ := currentMySQLActor(ctx, db)
	target := strings.TrimSpace(req.User) + "@" + strings.TrimSpace(req.Host)
	if strings.EqualFold(currentActor, target) {
		return errors.New("deleting the current mysql account is not allowed")
	}

	if err := execSecuritySQL(ctx, db, fmt.Sprintf("DROP USER %s", formatGrantee(req.User, req.Host))); err != nil {
		if compatErr := s.deletePrincipalCompat(ctx, db, req.Kind, req.User, req.Host, err); compatErr != nil {
			return compatErr
		}
	}

	if err := flushPrivileges(ctx, db); err != nil {
		return err
	}

	s.writeAuditLog(ctx, db, "delete-principal", req.User, req.Host, normalizePrincipalKind(req.Kind))
	return nil
}

func (s *SecurityService) ClonePrincipal(ctx context.Context, db *sql.DB, req model.CloneSecurityPrincipalRequest) error {
	targetKind := normalizePrincipalKind(req.TargetKind)
	sourceKind := targetKind
	if targetKind == model.SecurityPrincipalRole {
		sourceKind = model.SecurityPrincipalRole
	}

	sourceDetail, err := s.GetPrincipalDetail(ctx, db, req.SourceUser, req.SourceHost, sourceKind)
	if err != nil {
		return err
	}

	createReq := model.UpsertSecurityPrincipalRequest{
		User:             req.TargetUser,
		Host:             req.TargetHost,
		Kind:             targetKind,
		Password:         req.Password,
		PasswordChanged:  strings.TrimSpace(req.Password) != "",
		Locked:           sourceDetail.Locked,
		PasswordExpired:  sourceDetail.PasswordExpired,
		GlobalPrivileges: sourceDetail.GlobalPrivileges,
		SchemaPrivileges: sourceDetail.SchemaPrivileges,
		TablePrivileges:  sourceDetail.TablePrivileges,
		ColumnPrivileges: sourceDetail.ColumnPrivileges,
		Roles:            sourceDetail.Roles,
	}

	if err := s.CreatePrincipal(ctx, db, createReq); err != nil {
		return err
	}

	s.writeAuditLog(ctx, db, "clone-principal", req.TargetUser, req.TargetHost, targetKind)
	return nil
}

func (s *SecurityService) RevokeAll(ctx context.Context, db *sql.DB, req model.RevokeAllSecurityPrincipalRequest) error {
	capabilities, _, err := s.getCapabilities(ctx, db)
	if err != nil {
		return err
	}

	if err := s.revokeAllPrivileges(ctx, db, capabilities, req.User, req.Host); err != nil {
		return err
	}

	if err := flushPrivileges(ctx, db); err != nil {
		return err
	}

	s.writeAuditLog(ctx, db, "revoke-all", req.User, req.Host, normalizePrincipalKind(req.Kind))
	return nil
}

func (s *SecurityService) getCapabilities(ctx context.Context, db *sql.DB) (model.SecurityCapabilities, mysqlUserFeatures, error) {
	version := ""
	if err := db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err != nil {
		return model.SecurityCapabilities{}, mysqlUserFeatures{}, err
	}

	features, err := inspectMySQLUserFeatures(ctx, db)
	if err != nil {
		return model.SecurityCapabilities{}, mysqlUserFeatures{}, err
	}

	return model.SecurityCapabilities{
		Version:       version,
		SupportsRoles: features.hasIsRole,
	}, features, nil
}

func inspectMySQLUserFeatures(ctx context.Context, db *sql.DB) (mysqlUserFeatures, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT COLUMN_NAME
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = 'mysql' AND TABLE_NAME = 'user'
	`)
	if err != nil {
		return mysqlUserFeatures{}, err
	}
	defer rows.Close()

	features := mysqlUserFeatures{}
	for rows.Next() {
		var name string
		if scanErr := rows.Scan(&name); scanErr != nil {
			return mysqlUserFeatures{}, scanErr
		}

		switch strings.ToLower(name) {
		case "plugin":
			features.hasPlugin = true
		case "account_locked":
			features.hasAccountLocked = true
		case "password_expired":
			features.hasPasswordExpired = true
		case "is_role":
			features.hasIsRole = true
		}
	}

	return features, rows.Err()
}

func (s *SecurityService) listPrincipals(ctx context.Context, db *sql.DB, features mysqlUserFeatures, kind model.SecurityPrincipalKind) ([]model.SecurityPrincipalSummary, error) {
	query := []string{
		"SELECT User, Host,",
		sqlFeatureExpr(features.hasPlugin, "plugin", "''") + " AS plugin,",
		sqlFeatureExpr(features.hasAccountLocked, "account_locked", "'N'") + " AS account_locked,",
		sqlFeatureExpr(features.hasPasswordExpired, "password_expired", "'N'") + " AS password_expired,",
		sqlFeatureExpr(features.hasIsRole, "is_role", "'N'") + " AS is_role",
		"FROM mysql.user",
	}

	switch kind {
	case model.SecurityPrincipalRole:
		if !features.hasIsRole {
			return []model.SecurityPrincipalSummary{}, nil
		}
		query = append(query, "WHERE is_role = 'Y'")
	default:
		if features.hasIsRole {
			query = append(query, "WHERE is_role = 'N'")
		}
	}

	query = append(query, "ORDER BY User, Host")
	rows, err := db.QueryContext(ctx, strings.Join(query, " "))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.SecurityPrincipalSummary, 0)
	for rows.Next() {
		var user, host, plugin, accountLocked, passwordExpired, isRole string
		if scanErr := rows.Scan(&user, &host, &plugin, &accountLocked, &passwordExpired, &isRole); scanErr != nil {
			return nil, scanErr
		}

		summary := model.SecurityPrincipalSummary{
			User:            user,
			Host:            host,
			Kind:            kind,
			Locked:          normalizeMysqlFlag(accountLocked),
			PasswordExpired: normalizeMysqlFlag(passwordExpired),
			Plugin:          plugin,
		}

		privSummary, summaryErr := s.buildPrivilegeSummary(ctx, db, user, host)
		if summaryErr == nil {
			summary.PrivilegeSummary = privSummary
		}
		privDetails, detailsErr := s.buildPrivilegeDetails(ctx, db, user, host)
		if detailsErr == nil {
			summary.PrivilegeDetails = privDetails
		}

		items = append(items, summary)
	}

	return items, rows.Err()
}

func (s *SecurityService) loadPrincipalBase(ctx context.Context, db *sql.DB, features mysqlUserFeatures, user, host string, kind model.SecurityPrincipalKind) (model.SecurityPrincipalDetail, error) {
	query := []string{
		"SELECT User, Host,",
		sqlFeatureExpr(features.hasPlugin, "plugin", "''") + " AS plugin,",
		sqlFeatureExpr(features.hasAccountLocked, "account_locked", "'N'") + " AS account_locked,",
		sqlFeatureExpr(features.hasPasswordExpired, "password_expired", "'N'") + " AS password_expired,",
		sqlFeatureExpr(features.hasIsRole, "is_role", "'N'") + " AS is_role",
		"FROM mysql.user WHERE User = ? AND Host = ?",
	}

	if features.hasIsRole {
		if kind == model.SecurityPrincipalRole {
			query = append(query, "AND is_role = 'Y'")
		} else {
			query = append(query, "AND is_role = 'N'")
		}
	}

	var (
		foundUser        string
		foundHost        string
		plugin           string
		accountLocked    string
		passwordExpired  string
		isRole           string
	)

	if err := db.QueryRowContext(ctx, strings.Join(query, " "), user, host).Scan(&foundUser, &foundHost, &plugin, &accountLocked, &passwordExpired, &isRole); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.SecurityPrincipalDetail{}, errors.New("security principal not found")
		}
		return model.SecurityPrincipalDetail{}, err
	}

	actualKind := normalizePrincipalKind(kind)
	if features.hasIsRole && normalizeMysqlFlag(isRole) {
		actualKind = model.SecurityPrincipalRole
	}

	return model.SecurityPrincipalDetail{
		User:            foundUser,
		Host:            foundHost,
		Kind:            actualKind,
		Locked:          normalizeMysqlFlag(accountLocked),
		PasswordExpired: normalizeMysqlFlag(passwordExpired),
		Plugin:          plugin,
	}, nil
}

func (s *SecurityService) listGlobalPrivileges(ctx context.Context, db *sql.DB, grantee string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT PRIVILEGE_TYPE
		FROM information_schema.USER_PRIVILEGES
		WHERE GRANTEE = ?
		ORDER BY PRIVILEGE_TYPE
	`, grantee)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]string, 0)
	for rows.Next() {
		var privilege string
		if scanErr := rows.Scan(&privilege); scanErr != nil {
			return nil, scanErr
		}
		items = append(items, privilege)
	}
	return items, rows.Err()
}

func (s *SecurityService) listSchemaPrivileges(ctx context.Context, db *sql.DB, grantee string) ([]model.SecurityScopePrivileges, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT TABLE_SCHEMA, PRIVILEGE_TYPE
		FROM information_schema.SCHEMA_PRIVILEGES
		WHERE GRANTEE = ?
		ORDER BY TABLE_SCHEMA, PRIVILEGE_TYPE
	`, grantee)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grouped := make(map[string][]string)
	for rows.Next() {
		var database, privilege string
		if scanErr := rows.Scan(&database, &privilege); scanErr != nil {
			return nil, scanErr
		}
		grouped[database] = append(grouped[database], privilege)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	items := make([]model.SecurityScopePrivileges, 0, len(grouped))
	for database, privileges := range grouped {
		items = append(items, model.SecurityScopePrivileges{
			Database:   database,
			Privileges: uniqueSortedStrings(privileges),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Database < items[j].Database })
	return items, nil
}

func (s *SecurityService) listTablePrivileges(ctx context.Context, db *sql.DB, grantee string) ([]model.SecurityScopePrivileges, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT TABLE_SCHEMA, TABLE_NAME, PRIVILEGE_TYPE
		FROM information_schema.TABLE_PRIVILEGES
		WHERE GRANTEE = ?
		ORDER BY TABLE_SCHEMA, TABLE_NAME, PRIVILEGE_TYPE
	`, grantee)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grouped := make(map[string]*model.SecurityScopePrivileges)
	for rows.Next() {
		var database, tableName, privilege string
		if scanErr := rows.Scan(&database, &tableName, &privilege); scanErr != nil {
			return nil, scanErr
		}
		key := database + "." + tableName
		item, ok := grouped[key]
		if !ok {
			item = &model.SecurityScopePrivileges{Database: database, Table: tableName}
			grouped[key] = item
		}
		item.Privileges = append(item.Privileges, privilege)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	items := make([]model.SecurityScopePrivileges, 0, len(grouped))
	for _, item := range grouped {
		item.Privileges = uniqueSortedStrings(item.Privileges)
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Database == items[j].Database {
			return items[i].Table < items[j].Table
		}
		return items[i].Database < items[j].Database
	})
	return items, nil
}

func (s *SecurityService) listColumnPrivileges(ctx context.Context, db *sql.DB, grantee string) ([]model.SecurityScopePrivileges, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT TABLE_SCHEMA, TABLE_NAME, COLUMN_NAME, PRIVILEGE_TYPE
		FROM information_schema.COLUMN_PRIVILEGES
		WHERE GRANTEE = ?
		ORDER BY TABLE_SCHEMA, TABLE_NAME, COLUMN_NAME, PRIVILEGE_TYPE
	`, grantee)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.SecurityScopePrivileges, 0)
	for rows.Next() {
		var database, tableName, columnName, privilege string
		if scanErr := rows.Scan(&database, &tableName, &columnName, &privilege); scanErr != nil {
			return nil, scanErr
		}
		items = append(items, model.SecurityScopePrivileges{
			Database:   database,
			Table:      tableName,
			Column:     columnName,
			Privileges: []string{privilege},
		})
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(items) > 0 {
		return items, nil
	}

	user, host := splitGrantee(grantee)
	fallbackRows, fallbackErr := db.QueryContext(ctx, `
		SELECT Db, Table_name, Column_name, Column_priv
		FROM mysql.columns_priv
		WHERE User = ? AND Host = ?
		ORDER BY Db, Table_name, Column_name
	`, user, host)
	if fallbackErr != nil {
		return items, nil
	}
	defer fallbackRows.Close()

	for fallbackRows.Next() {
		var database, tableName, columnName, privilegeSet string
		if scanErr := fallbackRows.Scan(&database, &tableName, &columnName, &privilegeSet); scanErr != nil {
			return nil, scanErr
		}
		for _, privilege := range strings.Split(privilegeSet, ",") {
			normalized := strings.ToUpper(strings.TrimSpace(privilege))
			if normalized == "" {
				continue
			}
			items = append(items, model.SecurityScopePrivileges{
				Database:   database,
				Table:      tableName,
				Column:     columnName,
				Privileges: []string{normalized},
			})
		}
	}
	return items, fallbackRows.Err()
}

func (s *SecurityService) listGrantedRoles(ctx context.Context, db *sql.DB, user, host string) ([]model.SecurityPrincipalRef, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT FROM_USER, FROM_HOST
		FROM mysql.role_edges
		WHERE TO_USER = ? AND TO_HOST = ?
		ORDER BY FROM_USER, FROM_HOST
	`, user, host)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.SecurityPrincipalRef, 0)
	for rows.Next() {
		var roleUser, roleHost string
		if scanErr := rows.Scan(&roleUser, &roleHost); scanErr != nil {
			return nil, scanErr
		}
		items = append(items, model.SecurityPrincipalRef{
			User: roleUser,
			Host: roleHost,
		})
	}
	return items, rows.Err()
}

func (s *SecurityService) showGrantStatements(ctx context.Context, db *sql.DB, user, host string) ([]string, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("SHOW GRANTS FOR %s", formatGrantee(user, host)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	items := make([]string, 0)
	for rows.Next() {
		values := make([]sql.RawBytes, len(columns))
		scans := make([]any, len(columns))
		for index := range values {
			scans[index] = &values[index]
		}
		if scanErr := rows.Scan(scans...); scanErr != nil {
			return nil, scanErr
		}
		if len(values) > 0 {
			items = append(items, string(values[0]))
		}
	}
	return items, rows.Err()
}

func (s *SecurityService) buildPrivilegeSummary(ctx context.Context, db *sql.DB, user, host string) (string, error) {
	grantee := formatGrantee(user, host)
	counts := make([]int, 4)
	queries := []string{
		"SELECT COUNT(*) FROM information_schema.USER_PRIVILEGES WHERE GRANTEE = ?",
		"SELECT COUNT(*) FROM information_schema.SCHEMA_PRIVILEGES WHERE GRANTEE = ?",
		"SELECT COUNT(*) FROM information_schema.TABLE_PRIVILEGES WHERE GRANTEE = ?",
		"SELECT COUNT(*) FROM information_schema.COLUMN_PRIVILEGES WHERE GRANTEE = ?",
	}

	for index, query := range queries {
		if err := db.QueryRowContext(ctx, query, grantee).Scan(&counts[index]); err != nil {
			return "", err
		}
	}

	if counts[3] == 0 {
		user, host := splitGrantee(grantee)
		_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mysql.columns_priv WHERE User = ? AND Host = ?", user, host).Scan(&counts[3])
	}

	return fmt.Sprintf("Global %d / Schema %d / Table %d / Column %d", counts[0], counts[1], counts[2], counts[3]), nil
}

func (s *SecurityService) buildPrivilegeDetails(ctx context.Context, db *sql.DB, user, host string) (string, error) {
	grantee := formatGrantee(user, host)
	privileges := make([]string, 0, 32)

	globalPrivileges, err := s.listGlobalPrivileges(ctx, db, grantee)
	if err == nil {
		privileges = append(privileges, globalPrivileges...)
	}

	schemaPrivileges, err := s.listSchemaPrivileges(ctx, db, grantee)
	if err == nil {
		for _, item := range schemaPrivileges {
			privileges = append(privileges, item.Privileges...)
		}
	}

	tablePrivileges, err := s.listTablePrivileges(ctx, db, grantee)
	if err == nil {
		for _, item := range tablePrivileges {
			privileges = append(privileges, item.Privileges...)
		}
	}

	columnPrivileges, err := s.listColumnPrivileges(ctx, db, grantee)
	if err == nil {
		for _, item := range columnPrivileges {
			privileges = append(privileges, item.Privileges...)
		}
	}

	grantStatements, grantErr := s.showGrantStatements(ctx, db, user, host)
	if grantErr == nil {
		privileges = append(privileges, parsePrivilegesFromGrantStatements(grantStatements)...)
	}

	normalized := normalizePrivilegeNames(privileges)
	if len(normalized) == 0 {
		return "USAGE", nil
	}

	if len(normalized) > 1 {
		filtered := make([]string, 0, len(normalized))
		for _, item := range normalized {
			if item == "USAGE" {
				continue
			}
			filtered = append(filtered, item)
		}
		if len(filtered) > 0 {
			normalized = filtered
		}
	}

	return strings.Join(normalized, ", "), nil
}

func (s *SecurityService) applyPrincipalSecurity(ctx context.Context, db *sql.DB, capabilities model.SecurityCapabilities, req model.UpsertSecurityPrincipalRequest, isUpdate bool) error {
	kind := normalizePrincipalKind(req.Kind)
	grantee := formatGrantee(req.User, req.Host)

	if kind == model.SecurityPrincipalUser {
		if isUpdate && req.PasswordChanged {
			if err := execSecuritySQL(ctx, db, fmt.Sprintf("ALTER USER %s IDENTIFIED BY %s", grantee, quoteString(req.Password))); err != nil {
				if !isGrantTableCompatError(err) {
					return err
				}
			}
		}

		if err := s.applyUserState(ctx, db, capabilities, req); err != nil {
			return err
		}
	}

	if err := s.revokeAllPrivileges(ctx, db, capabilities, req.User, req.Host); err != nil {
		return err
	}

	useCompatGrantTables := false
	if len(req.GlobalPrivileges) > 0 {
		if err := execSecuritySQL(ctx, db, fmt.Sprintf("GRANT %s ON *.* TO %s", joinPrivileges(req.GlobalPrivileges), grantee)); err != nil {
			if !isGrantTableCompatError(err) {
				return err
			}
			useCompatGrantTables = true
		}
	}

	if !useCompatGrantTables {
		for _, scope := range req.SchemaPrivileges {
			database := strings.TrimSpace(scope.Database)
			if database == "" || len(scope.Privileges) == 0 {
				continue
			}
			if err := execSecuritySQL(ctx, db, fmt.Sprintf("GRANT %s ON %s.* TO %s", joinPrivileges(scope.Privileges), quoteIdentifier(database), grantee)); err != nil {
				if !isGrantTableCompatError(err) {
					return err
				}
				useCompatGrantTables = true
				break
			}
		}
	}

	if !useCompatGrantTables {
		for _, scope := range req.TablePrivileges {
			database := strings.TrimSpace(scope.Database)
			tableName := strings.TrimSpace(scope.Table)
			if database == "" || tableName == "" || len(scope.Privileges) == 0 {
				continue
			}
			if err := execSecuritySQL(ctx, db, fmt.Sprintf("GRANT %s ON %s.%s TO %s", joinPrivileges(scope.Privileges), quoteIdentifier(database), quoteIdentifier(tableName), grantee)); err != nil {
				if !isGrantTableCompatError(err) {
					return err
				}
				useCompatGrantTables = true
				break
			}
		}
	}

	if !useCompatGrantTables && len(req.ColumnPrivileges) > 0 {
		if err := s.applyColumnPrivileges(ctx, db, grantee, req.ColumnPrivileges); err != nil {
			if !isGrantTableCompatError(err) {
				return err
			}
			useCompatGrantTables = true
		}
	}

	if capabilities.SupportsRoles && len(req.Roles) > 0 {
		roleTargets := make([]string, 0, len(req.Roles))
		for _, role := range req.Roles {
			if strings.TrimSpace(role.User) == "" || strings.TrimSpace(role.Host) == "" {
				continue
			}
			roleTargets = append(roleTargets, formatGrantee(role.User, role.Host))
		}
		if len(roleTargets) > 0 {
			if err := execSecuritySQL(ctx, db, fmt.Sprintf("GRANT %s TO %s", strings.Join(roleTargets, ", "), grantee)); err != nil {
				return err
			}
		}
	}

	if useCompatGrantTables {
		if compatErr := s.applyPrincipalSecurityCompat(ctx, db, req); compatErr != nil {
			return compatErr
		}
	}

	return flushPrivileges(ctx, db)
}

func (s *SecurityService) applyUserState(ctx context.Context, db *sql.DB, capabilities model.SecurityCapabilities, req model.UpsertSecurityPrincipalRequest) error {
	grantee := formatGrantee(req.User, req.Host)

	lockKeyword := "UNLOCK"
	if req.Locked {
		lockKeyword = "LOCK"
	}
	if err := execSecuritySQL(ctx, db, fmt.Sprintf("ALTER USER %s ACCOUNT %s", grantee, lockKeyword)); err != nil && !strings.Contains(strings.ToLower(err.Error()), "syntax") {
		if !isGrantTableCompatError(err) {
			return err
		}
	}

	expireKeyword := "PASSWORD EXPIRE NEVER"
	if req.PasswordExpired {
		expireKeyword = "PASSWORD EXPIRE"
	}
	if err := execSecuritySQL(ctx, db, fmt.Sprintf("ALTER USER %s %s", grantee, expireKeyword)); err != nil && !strings.Contains(strings.ToLower(err.Error()), "syntax") {
		if !isGrantTableCompatError(err) {
			return err
		}
	}

	if compatErr := s.applyUserStateCompat(ctx, db, req); compatErr != nil {
		return compatErr
	}

	_ = capabilities
	return nil
}

func (s *SecurityService) revokeAllPrivileges(ctx context.Context, db *sql.DB, capabilities model.SecurityCapabilities, user, host string) error {
	grantee := formatGrantee(user, host)
	if err := execSecuritySQL(ctx, db, fmt.Sprintf("REVOKE ALL PRIVILEGES, GRANT OPTION FROM %s", grantee)); err != nil {
		lowered := strings.ToLower(err.Error())
		if isGrantTableCompatError(err) {
			return s.revokeAllPrivilegesCompat(ctx, db, user, host)
		}
		if !strings.Contains(lowered, "there is no such grant defined") && !strings.Contains(lowered, "not allowed") {
			return err
		}
	}

	if capabilities.SupportsRoles {
		roles, err := s.listGrantedRoles(ctx, db, user, host)
		if err == nil && len(roles) > 0 {
			roleTargets := make([]string, 0, len(roles))
			for _, role := range roles {
				roleTargets = append(roleTargets, formatGrantee(role.User, role.Host))
			}
			if len(roleTargets) > 0 {
				if revokeErr := execSecuritySQL(ctx, db, fmt.Sprintf("REVOKE %s FROM %s", strings.Join(roleTargets, ", "), grantee)); revokeErr != nil {
					return revokeErr
				}
			}
		}
	}

	return nil
}

func (s *SecurityService) applyColumnPrivileges(ctx context.Context, db *sql.DB, grantee string, items []model.SecurityScopePrivileges) error {
	type groupedColumnPrivileges struct {
		database string
		table    string
		values   map[string][]string
	}

	grouped := make(map[string]*groupedColumnPrivileges)
	for _, item := range items {
		database := strings.TrimSpace(item.Database)
		tableName := strings.TrimSpace(item.Table)
		columnName := strings.TrimSpace(item.Column)
		if database == "" || tableName == "" || columnName == "" || len(item.Privileges) == 0 {
			continue
		}
		key := database + "." + tableName
		entry, ok := grouped[key]
		if !ok {
			entry = &groupedColumnPrivileges{
				database: database,
				table:    tableName,
				values:   make(map[string][]string),
			}
			grouped[key] = entry
		}

		for _, privilege := range item.Privileges {
			entry.values[strings.ToUpper(strings.TrimSpace(privilege))] = append(entry.values[strings.ToUpper(strings.TrimSpace(privilege))], columnName)
		}
	}

	keys := make([]string, 0, len(grouped))
	for key := range grouped {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		entry := grouped[key]
		fragments := make([]string, 0, len(entry.values))
		privilegeNames := make([]string, 0, len(entry.values))
		for privilegeName := range entry.values {
			privilegeNames = append(privilegeNames, privilegeName)
		}
		sort.Strings(privilegeNames)
		for _, privilegeName := range privilegeNames {
			columns := uniqueSortedStrings(entry.values[privilegeName])
			quotedColumns := make([]string, 0, len(columns))
			for _, columnName := range columns {
				quotedColumns = append(quotedColumns, quoteIdentifier(columnName))
			}
			fragments = append(fragments, fmt.Sprintf("%s (%s)", privilegeName, strings.Join(quotedColumns, ", ")))
		}

		if len(fragments) == 0 {
			continue
		}

		statement := fmt.Sprintf(
			"GRANT %s ON %s.%s TO %s",
			strings.Join(fragments, ", "),
			quoteIdentifier(entry.database),
			quoteIdentifier(entry.table),
			grantee,
		)
		if err := execSecuritySQL(ctx, db, statement); err != nil {
			return err
		}
	}

	return nil
}

func (s *SecurityService) createPrincipalCompat(ctx context.Context, db *sql.DB, req model.UpsertSecurityPrincipalRequest, cause error) error {
	if !isGrantTableCompatError(cause) {
		return cause
	}
	if normalizePrincipalKind(req.Kind) != model.SecurityPrincipalUser {
		return cause
	}

	stmt := `
		INSERT INTO mysql.user (
			Host, User, ssl_cipher, x509_issuer, x509_subject,
			max_questions, max_updates, max_connections, max_user_connections,
			plugin, authentication_string, password_expired, account_locked
		) VALUES (?, ?, '', '', '', 0, 0, 0, 0, ?, ?, ?, ?)
	`
	pluginName, authString := buildCompatAuth(req.Password)
	if _, err := db.ExecContext(
		ctx,
		stmt,
		strings.TrimSpace(req.Host),
		strings.TrimSpace(req.User),
		pluginName,
		authString,
		boolToMysqlFlag(req.PasswordExpired),
		boolToMysqlFlag(req.Locked),
	); err != nil {
		return fmt.Errorf("compat create user fallback failed: %w", err)
	}

	return flushPrivileges(ctx, db)
}

func (s *SecurityService) renamePrincipalCompat(ctx context.Context, db *sql.DB, kind model.SecurityPrincipalKind, originalUser, originalHost, nextUser, nextHost string, cause error) error {
	if !isGrantTableCompatError(cause) {
		return cause
	}
	if normalizePrincipalKind(kind) != model.SecurityPrincipalUser {
		return cause
	}

	updates := []string{
		"UPDATE mysql.user SET User = ?, Host = ? WHERE User = ? AND Host = ?",
		"UPDATE mysql.db SET User = ?, Host = ? WHERE User = ? AND Host = ?",
		"UPDATE mysql.tables_priv SET User = ?, Host = ? WHERE User = ? AND Host = ?",
		"UPDATE mysql.columns_priv SET User = ?, Host = ? WHERE User = ? AND Host = ?",
		"UPDATE mysql.role_edges SET TO_USER = ?, TO_HOST = ? WHERE TO_USER = ? AND TO_HOST = ?",
		"UPDATE mysql.default_roles SET USER = ?, HOST = ? WHERE USER = ? AND HOST = ?",
	}
	for _, stmt := range updates {
		if _, err := db.ExecContext(ctx, stmt, nextUser, nextHost, originalUser, originalHost); err != nil {
			return fmt.Errorf("compat rename principal fallback failed: %w", err)
		}
	}
	return flushPrivileges(ctx, db)
}

func (s *SecurityService) deletePrincipalCompat(ctx context.Context, db *sql.DB, kind model.SecurityPrincipalKind, user, host string, cause error) error {
	if !isGrantTableCompatError(cause) {
		return cause
	}
	if normalizePrincipalKind(kind) != model.SecurityPrincipalUser {
		return cause
	}

	deletes := []string{
		"DELETE FROM mysql.default_roles WHERE USER = ? AND HOST = ?",
		"DELETE FROM mysql.role_edges WHERE TO_USER = ? AND TO_HOST = ?",
		"DELETE FROM mysql.role_edges WHERE FROM_USER = ? AND FROM_HOST = ?",
		"DELETE FROM mysql.columns_priv WHERE User = ? AND Host = ?",
		"DELETE FROM mysql.tables_priv WHERE User = ? AND Host = ?",
		"DELETE FROM mysql.db WHERE User = ? AND Host = ?",
		"DELETE FROM mysql.user WHERE User = ? AND Host = ?",
	}
	for _, stmt := range deletes {
		if _, err := db.ExecContext(ctx, stmt, user, host); err != nil {
			return fmt.Errorf("compat delete principal fallback failed: %w", err)
		}
	}
	return flushPrivileges(ctx, db)
}

func (s *SecurityService) applyPrincipalSecurityCompat(ctx context.Context, db *sql.DB, req model.UpsertSecurityPrincipalRequest) error {
	if err := s.applyGlobalPrivilegesCompat(ctx, db, req.User, req.Host, req.GlobalPrivileges); err != nil {
		return err
	}
	if err := s.applySchemaPrivilegesCompat(ctx, db, req.User, req.Host, req.SchemaPrivileges); err != nil {
		return err
	}
	if err := s.applyTablePrivilegesCompat(ctx, db, req.User, req.Host, req.TablePrivileges); err != nil {
		return err
	}
	if err := s.applyColumnPrivilegesCompat(ctx, db, req.User, req.Host, req.ColumnPrivileges); err != nil {
		return err
	}
	return flushPrivileges(ctx, db)
}

func (s *SecurityService) applyUserStateCompat(ctx context.Context, db *sql.DB, req model.UpsertSecurityPrincipalRequest) error {
	parts := []string{
		"UPDATE mysql.user SET account_locked = ?, password_expired = ?",
	}
	args := []any{boolToMysqlFlag(req.Locked), boolToMysqlFlag(req.PasswordExpired)}

	if strings.TrimSpace(req.Password) != "" && req.PasswordChanged {
		pluginName, authString := buildCompatAuth(req.Password)
		parts = append(parts, ", plugin = ?, authentication_string = ?")
		args = append(args, pluginName, authString)
	}

	parts = append(parts, "WHERE User = ? AND Host = ?")
	args = append(args, req.User, req.Host)

	if _, err := db.ExecContext(ctx, strings.Join(parts, " "), args...); err != nil {
		return fmt.Errorf("compat update user state failed: %w", err)
	}
	return nil
}

func (s *SecurityService) revokeAllPrivilegesCompat(ctx context.Context, db *sql.DB, user, host string) error {
	globalColumns := uniqueSortedStrings(mapValues(globalPrivilegeColumnMap))
	if len(globalColumns) > 0 {
		setClauses := make([]string, 0, len(globalColumns))
		args := make([]any, 0, len(globalColumns)+2)
		for _, column := range globalColumns {
			setClauses = append(setClauses, quoteIdentifier(column)+" = 'N'")
		}
		args = append(args, user, host)
		statement := "UPDATE mysql.user SET " + strings.Join(setClauses, ", ") + " WHERE User = ? AND Host = ?"
		if _, err := db.ExecContext(ctx, statement, args...); err != nil {
			return fmt.Errorf("compat revoke global privileges failed: %w", err)
		}
	}

	deletes := []string{
		"DELETE FROM mysql.default_roles WHERE USER = ? AND HOST = ?",
		"DELETE FROM mysql.role_edges WHERE TO_USER = ? AND TO_HOST = ?",
		"DELETE FROM mysql.columns_priv WHERE User = ? AND Host = ?",
		"DELETE FROM mysql.tables_priv WHERE User = ? AND Host = ?",
		"DELETE FROM mysql.db WHERE User = ? AND Host = ?",
	}
	for _, stmt := range deletes {
		if _, err := db.ExecContext(ctx, stmt, user, host); err != nil {
			return fmt.Errorf("compat revoke privilege rows failed: %w", err)
		}
	}
	return nil
}

func (s *SecurityService) applyGlobalPrivilegesCompat(ctx context.Context, db *sql.DB, user, host string, privileges []string) error {
	enabled := make(map[string]struct{}, len(privileges))
	for _, privilege := range uniqueSortedStrings(privileges) {
		if column, ok := globalPrivilegeColumnMap[strings.ToUpper(privilege)]; ok {
			enabled[column] = struct{}{}
		}
	}

	columns := uniqueSortedStrings(mapValues(globalPrivilegeColumnMap))
	assignments := make([]string, 0, len(columns))
	for _, column := range columns {
		value := "'N'"
		if _, ok := enabled[column]; ok {
			value = "'Y'"
		}
		assignments = append(assignments, quoteIdentifier(column)+" = "+value)
	}
	statement := "UPDATE mysql.user SET " + strings.Join(assignments, ", ") + " WHERE User = ? AND Host = ?"
	if _, err := db.ExecContext(ctx, statement, user, host); err != nil {
		return fmt.Errorf("compat apply global privileges failed: %w", err)
	}
	return nil
}

func (s *SecurityService) applySchemaPrivilegesCompat(ctx context.Context, db *sql.DB, user, host string, scopes []model.SecurityScopePrivileges) error {
	if _, err := db.ExecContext(ctx, "DELETE FROM mysql.db WHERE User = ? AND Host = ?", user, host); err != nil {
		return fmt.Errorf("compat clear schema privileges failed: %w", err)
	}

	for _, scope := range scopes {
		database := strings.TrimSpace(scope.Database)
		if database == "" {
			continue
		}

		assignments := schemaPrivilegeAssignments(scope.Privileges)
		if len(assignments.columns) == 0 {
			continue
		}

		stmt := "INSERT INTO mysql.db (Host, Db, User, " + strings.Join(assignments.columns, ", ") + ") VALUES (?, ?, ?, " + strings.Join(assignments.values, ", ") + ")"
		if _, err := db.ExecContext(ctx, stmt, host, database, user); err != nil {
			return fmt.Errorf("compat apply schema privileges failed: %w", err)
		}
	}
	return nil
}

func (s *SecurityService) applyTablePrivilegesCompat(ctx context.Context, db *sql.DB, user, host string, scopes []model.SecurityScopePrivileges) error {
	if _, err := db.ExecContext(ctx, "DELETE FROM mysql.tables_priv WHERE User = ? AND Host = ?", user, host); err != nil {
		return fmt.Errorf("compat clear table privileges failed: %w", err)
	}

	for _, scope := range scopes {
		database := strings.TrimSpace(scope.Database)
		tableName := strings.TrimSpace(scope.Table)
		items := normalizePrivilegeSet(scope.Privileges, tablePrivilegeSetMap)
		if database == "" || tableName == "" || len(items) == 0 {
			continue
		}

		tablePriv := quoteString(strings.Join(items, ","))
		if _, err := db.ExecContext(
			ctx,
			"INSERT INTO mysql.tables_priv (Host, Db, User, Table_name, Grantor, Table_priv, Column_priv) VALUES (?, ?, ?, ?, ?, "+tablePriv+", '')",
			host,
			database,
			user,
			tableName,
			currentGrantorFallback(user, host),
		); err != nil {
			return fmt.Errorf("compat apply table privileges failed: %w", err)
		}
	}
	return nil
}

func (s *SecurityService) applyColumnPrivilegesCompat(ctx context.Context, db *sql.DB, user, host string, scopes []model.SecurityScopePrivileges) error {
	if _, err := db.ExecContext(ctx, "DELETE FROM mysql.columns_priv WHERE User = ? AND Host = ?", user, host); err != nil {
		return fmt.Errorf("compat clear column privileges failed: %w", err)
	}

	for _, scope := range scopes {
		database := strings.TrimSpace(scope.Database)
		tableName := strings.TrimSpace(scope.Table)
		columnName := strings.TrimSpace(scope.Column)
		items := normalizePrivilegeSet(scope.Privileges, columnPrivilegeSetMap)
		if database == "" || tableName == "" || columnName == "" || len(items) == 0 {
			continue
		}

		columnPriv := quoteString(strings.Join(items, ","))
		if _, err := db.ExecContext(
			ctx,
			"INSERT INTO mysql.columns_priv (Host, Db, User, Table_name, Column_name, Column_priv) VALUES (?, ?, ?, ?, ?, "+columnPriv+")",
			host,
			database,
			user,
			tableName,
			columnName,
		); err != nil {
			return fmt.Errorf("compat apply column privileges failed: %w", err)
		}
	}
	return nil
}

func (s *SecurityService) writeAuditLog(ctx context.Context, db *sql.DB, action, user, host string, kind model.SecurityPrincipalKind) {
	actor, err := currentMySQLActor(ctx, db)
	if err != nil {
		actor = "unknown"
	}

	log.Printf("audit:mysql-security action=%s actor=%s kind=%s target=%s", action, actor, kind, formatGrantee(user, host))
}

func currentMySQLActor(ctx context.Context, db *sql.DB) (string, error) {
	var actor string
	if err := db.QueryRowContext(ctx, "SELECT CURRENT_USER()").Scan(&actor); err != nil {
		return "", err
	}
	return actor, nil
}

func execSecuritySQL(ctx context.Context, db *sql.DB, statement string) error {
	if _, err := db.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("%s: %w", statement, err)
	}
	return nil
}

func flushPrivileges(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "FLUSH PRIVILEGES")
	return err
}

func sqlFeatureExpr(enabled bool, column, fallback string) string {
	if enabled {
		return column
	}
	return fallback
}

func normalizeMysqlFlag(value string) bool {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "Y", "YES", "1", "TRUE", "LOCKED":
		return true
	default:
		return false
	}
}

func normalizePrincipalKind(kind model.SecurityPrincipalKind) model.SecurityPrincipalKind {
	if kind == model.SecurityPrincipalRole {
		return model.SecurityPrincipalRole
	}
	return model.SecurityPrincipalUser
}

func formatGrantee(user, host string) string {
	return quoteString(strings.TrimSpace(user)) + "@" + quoteString(strings.TrimSpace(host))
}

func quoteString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func quoteIdentifier(name string) string {
	return "`" + strings.ReplaceAll(strings.TrimSpace(name), "`", "``") + "`"
}

func uniqueSortedStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	items := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		items = append(items, normalized)
	}
	sort.Strings(items)
	return items
}

func normalizePrivilegeNames(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	items := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.ToUpper(strings.TrimSpace(strings.ReplaceAll(value, "`", "")))
		if normalized == "" {
			continue
		}
		normalized = strings.Join(strings.Fields(normalized), " ")
		normalized = strings.TrimSuffix(normalized, " PRIVILEGES")
		normalized = strings.TrimSpace(normalized)
		if normalized == "" {
			continue
		}
		if normalized == "ALL" {
			normalized = "ALL PRIVILEGES"
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		items = append(items, normalized)
	}
	sort.Strings(items)
	return items
}

func parsePrivilegesFromGrantStatements(statements []string) []string {
	privileges := make([]string, 0, 16)
	for _, statement := range statements {
		trimmed := strings.TrimSpace(statement)
		upper := strings.ToUpper(trimmed)
		if !strings.HasPrefix(upper, "GRANT ") {
			continue
		}

		onIndex := strings.Index(upper, " ON ")
		if onIndex < 0 {
			continue
		}

		clause := strings.TrimSpace(trimmed[len("GRANT "):onIndex])
		if clause == "" {
			continue
		}

		for _, fragment := range splitGrantPrivilegeClause(clause) {
			fragment = strings.TrimSpace(fragment)
			if fragment == "" {
				continue
			}
			if bracketIndex := strings.Index(fragment, "("); bracketIndex >= 0 {
				fragment = strings.TrimSpace(fragment[:bracketIndex])
			}
			if fragment != "" {
				privileges = append(privileges, fragment)
			}
		}
	}
	return privileges
}

func splitGrantPrivilegeClause(clause string) []string {
	parts := make([]string, 0, 8)
	start := 0
	depth := 0
	for index, char := range clause {
		switch char {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				parts = append(parts, clause[start:index])
				start = index + 1
			}
		}
	}
	parts = append(parts, clause[start:])
	return parts
}

func joinPrivileges(privileges []string) string {
	items := uniqueSortedStrings(privileges)
	for index, item := range items {
		items[index] = strings.ToUpper(item)
	}
	return strings.Join(items, ", ")
}

func parseMySQLVersionMajor(version string) int {
	number := strings.TrimSpace(version)
	if index := strings.Index(number, "."); index > 0 {
		number = number[:index]
	}
	major, err := strconv.Atoi(number)
	if err != nil {
		return 0
	}
	return major
}

var globalPrivilegeColumnMap = map[string]string{
	"SELECT":             "Select_priv",
	"INSERT":             "Insert_priv",
	"UPDATE":             "Update_priv",
	"DELETE":             "Delete_priv",
	"CREATE":             "Create_priv",
	"DROP":               "Drop_priv",
	"ALTER":              "Alter_priv",
	"INDEX":              "Index_priv",
	"CREATE USER":        "Create_user_priv",
	"RELOAD":             "Reload_priv",
	"PROCESS":            "Process_priv",
	"SHOW DATABASES":     "Show_db_priv",
	"SUPER":              "Super_priv",
	"REPLICATION SLAVE":  "Repl_slave_priv",
	"REPLICATION CLIENT": "Repl_client_priv",
	"TRIGGER":            "Trigger_priv",
	"EVENT":              "Event_priv",
	"EXECUTE":            "Execute_priv",
	"CREATE VIEW":        "Create_view_priv",
	"SHOW VIEW":          "Show_view_priv",
	"CREATE ROUTINE":     "Create_routine_priv",
	"ALTER ROUTINE":      "Alter_routine_priv",
	"REFERENCES":         "References_priv",
}

var schemaPrivilegeColumnMap = map[string]string{
	"SELECT":         "Select_priv",
	"INSERT":         "Insert_priv",
	"UPDATE":         "Update_priv",
	"DELETE":         "Delete_priv",
	"CREATE":         "Create_priv",
	"DROP":           "Drop_priv",
	"ALTER":          "Alter_priv",
	"INDEX":          "Index_priv",
	"CREATE VIEW":    "Create_view_priv",
	"SHOW VIEW":      "Show_view_priv",
	"TRIGGER":        "Trigger_priv",
	"EVENT":          "Event_priv",
	"EXECUTE":        "Execute_priv",
	"CREATE ROUTINE": "Create_routine_priv",
	"ALTER ROUTINE":  "Alter_routine_priv",
}

var tablePrivilegeSetMap = map[string]string{
	"SELECT":     "Select",
	"INSERT":     "Insert",
	"UPDATE":     "Update",
	"DELETE":     "Delete",
	"CREATE":     "Create",
	"DROP":       "Drop",
	"ALTER":      "Alter",
	"INDEX":      "Index",
	"TRIGGER":    "Trigger",
	"REFERENCES": "References",
}

var columnPrivilegeSetMap = map[string]string{
	"SELECT":     "Select",
	"INSERT":     "Insert",
	"UPDATE":     "Update",
	"REFERENCES": "References",
}

type compatInsertAssignments struct {
	columns []string
	values  []string
}

func isGrantTableCompatError(err error) bool {
	if err == nil {
		return false
	}
	lowered := strings.ToLower(err.Error())
	return strings.Contains(lowered, "does not support system tables") && strings.Contains(lowered, "mysql.db")
}

func buildCompatAuth(password string) (string, string) {
	sum1 := sha1.Sum([]byte(password))
	sum2 := sha1.Sum(sum1[:])
	return "mysql_native_password", "*" + strings.ToUpper(hex.EncodeToString(sum2[:]))
}

func boolToMysqlFlag(value bool) string {
	if value {
		return "Y"
	}
	return "N"
}

func normalizePrivilegeSet(privileges []string, mapping map[string]string) []string {
	items := make([]string, 0, len(privileges))
	seen := make(map[string]struct{}, len(privileges))
	for _, privilege := range privileges {
		mapped, ok := mapping[strings.ToUpper(strings.TrimSpace(privilege))]
		if !ok {
			continue
		}
		if _, exists := seen[mapped]; exists {
			continue
		}
		seen[mapped] = struct{}{}
		items = append(items, mapped)
	}
	sort.Strings(items)
	return items
}

func schemaPrivilegeAssignments(privileges []string) compatInsertAssignments {
	items := normalizePrivilegeSet(privileges, mapPrivilegeColumns(schemaPrivilegeColumnMap))
	columns := make([]string, 0, len(items))
	values := make([]string, 0, len(items))
	for _, column := range items {
		columns = append(columns, column)
		values = append(values, "'Y'")
	}
	return compatInsertAssignments{
		columns: columns,
		values:  values,
	}
}

func mapPrivilegeColumns(source map[string]string) map[string]string {
	mapped := make(map[string]string, len(source))
	for privilege, column := range source {
		mapped[privilege] = column
	}
	return mapped
}

func mapValues(source map[string]string) []string {
	values := make([]string, 0, len(source))
	for _, value := range source {
		values = append(values, value)
	}
	return values
}

func currentGrantorFallback(user, host string) string {
	if strings.TrimSpace(user) == "" {
		return "root@localhost"
	}
	return user + "@" + host
}

func splitGrantee(grantee string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(grantee), "@", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return strings.Trim(parts[0], "'"), strings.Trim(parts[1], "'")
}
