package service

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"

	"devops-console-backend/internal/mysql/model"
)

type ConnectionManager struct {
	pool *DBPool
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		pool: NewDBPool(),
	}
}

func (m *ConnectionManager) Open(ctx context.Context, req model.OpenConnectionRequest) (string, error) {
	configs, err := buildCandidateConfigs(req)
	if err != nil {
		return "", err
	}

	var attemptErrors []string
	for _, cfg := range configs {
		db, openErr := sql.Open("mysql", cfg.FormatDSN())
		if openErr != nil {
			attemptErrors = append(attemptErrors, fmt.Sprintf("%s: open failed: %v", cfg.Addr, openErr))
			continue
		}

		configureDBPool(db)

		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		pingErr := db.PingContext(pingCtx)
		cancel()
		if pingErr != nil {
			_ = db.Close()
			attemptErrors = append(attemptErrors, fmt.Sprintf("%s: %v", cfg.Addr, pingErr))
			continue
		}

		if _, execErr := db.ExecContext(ctx, "SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci"); execErr != nil {
			_ = db.Close()
			attemptErrors = append(attemptErrors, fmt.Sprintf("%s: set names failed: %v", cfg.Addr, execErr))
			continue
		}

		token := uuid.NewString()
		m.pool.Store(token, DBSession{
			DB:      db,
			Profile: req,
		})
		return token, nil
	}

	if len(attemptErrors) == 0 {
		return "", errors.New("no mysql connection attempt was made")
	}

	return "", fmt.Errorf("ping mysql failed: %s", strings.Join(attemptErrors, "; "))
}

func (m *ConnectionManager) Close(token string) error {
	session, ok := m.pool.Load(token)
	if !ok {
		return errors.New("connection token not found")
	}

	m.pool.Delete(token)
	if err := session.DB.Close(); err != nil {
		return fmt.Errorf("close mysql connection failed: %w", err)
	}

	return nil
}

func (m *ConnectionManager) Get(token string) (*sql.DB, error) {
	session, ok := m.pool.Load(token)
	if !ok {
		return nil, errors.New("connection token not found")
	}

	return session.DB, nil
}

func (m *ConnectionManager) GetSession(token string) (*DBSession, error) {
	session, ok := m.pool.Load(token)
	if !ok {
		return nil, errors.New("connection token not found")
	}

	return &session, nil
}

func buildCandidateConfigs(req model.OpenConnectionRequest) ([]*mysql.Config, error) {
	base := &mysql.Config{
		User:                 req.Username,
		Passwd:               req.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", req.Host, req.Port),
		DBName:               req.Database,
		AllowNativePasswords: true,
		ParseTime:            true,
		Loc:                  time.Local,
		Params: map[string]string{
			"charset":   "utf8mb4",
			"collation": "utf8mb4_unicode_ci",
		},
	}

	for key, value := range req.Params {
		base.Params[key] = value
	}

	addresses := buildCandidateAddresses(req.Host, req.Port)
	configs := make([]*mysql.Config, 0, len(addresses))
	for _, addr := range addresses {
		cfg := *base
		cfg.Addr = addr
		configs = append(configs, &cfg)
	}

	return configs, nil
}

func configureDBPool(db *sql.DB) {
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)
}

func buildCandidateAddresses(host string, port int) []string {
	normalizedHost := normalizeMySQLHost(host)
	baseAddr := fmt.Sprintf("%s:%d", normalizedHost, port)
	addresses := []string{baseAddr}

	if isLocalMachineHost(normalizedHost) {
		addresses = append(
			addresses,
			fmt.Sprintf("127.0.0.1:%d", port),
			fmt.Sprintf("localhost:%d", port),
		)
	}

	if fallbackAddr, ok := resolveWSLHostFallback(normalizedHost, port); ok {
		addresses = append(addresses, fallbackAddr)
	}

	return uniqueAddresses(addresses)
}

func normalizeMySQLHost(host string) string {
	trimmed := strings.TrimSpace(host)
	switch trimmed {
	case "", "0.0.0.0", "::", "[::]":
		return "127.0.0.1"
	default:
		return trimmed
	}
}

func isLocalMachineHost(host string) bool {
	if host == "127.0.0.1" || host == "localhost" || host == "::1" || host == "[::1]" {
		return true
	}

	ip := net.ParseIP(strings.Trim(host, "[]"))
	if ip == nil {
		return false
	}

	if ip.IsLoopback() || ip.IsUnspecified() {
		return true
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}

	for _, addr := range addrs {
		network, ok := addr.(*net.IPNet)
		if !ok || network.IP == nil {
			continue
		}

		if network.IP.Equal(ip) {
			return true
		}
	}

	return false
}

func uniqueAddresses(addresses []string) []string {
	seen := make(map[string]struct{}, len(addresses))
	unique := make([]string, 0, len(addresses))
	for _, addr := range addresses {
		if addr == "" {
			continue
		}
		if _, ok := seen[addr]; ok {
			continue
		}
		seen[addr] = struct{}{}
		unique = append(unique, addr)
	}

	return unique
}

func resolveWSLHostFallback(host string, port int) (string, bool) {
	if host != "127.0.0.1" && host != "localhost" {
		return "", false
	}

	if !isRunningInWSL() {
		return "", false
	}

	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return "", false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "nameserver ") {
			continue
		}

		ip := strings.TrimSpace(strings.TrimPrefix(line, "nameserver "))
		if ip == "" {
			continue
		}

		return fmt.Sprintf("%s:%d", ip, port), true
	}

	return "", false
}

func isRunningInWSL() bool {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false
	}

	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}

