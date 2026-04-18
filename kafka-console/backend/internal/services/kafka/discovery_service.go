package kafka

import (
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"devops-console-backend/internal/dal"
	reqKafka "devops-console-backend/internal/dal/request/kafka"
	"devops-console-backend/internal/dal/response"

	"github.com/IBM/sarama"
	"github.com/xdg-go/scram"
)

func (s *Service) ScanKafkaNetwork(req reqKafka.DiscoveryScanRequest) ([]response.KafkaDiscoveryResultVO, error) {
	hosts, err := expandCIDR(req.CIDR)
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(req.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 2500 * time.Millisecond
	}
	concurrency := req.Concurrency
	if concurrency <= 0 {
		concurrency = 64
	}

	type job struct {
		ip   string
		port int
	}
	jobs := make(chan job)
	results := make(chan response.KafkaDiscoveryResultVO, len(hosts)*len(req.Ports))
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobs {
				results <- s.probeKafkaEndpoint(item.ip, item.port, timeout, req.Auth)
			}
		}()
	}

	go func() {
		for _, host := range hosts {
			for _, port := range req.Ports {
				jobs <- job{ip: host, port: port}
			}
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	list := make([]response.KafkaDiscoveryResultVO, 0, len(hosts)*len(req.Ports))
	for item := range results {
		list = append(list, item)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].LooksLikeKafka != list[j].LooksLikeKafka {
			return list[i].LooksLikeKafka
		}
		if list[i].IP == list[j].IP {
			return list[i].Port < list[j].Port
		}
		return list[i].IP < list[j].IP
	})
	return list, nil
}

func (s *Service) ImportDiscoveredKafka(req reqKafka.DiscoveryImportRequest) (*response.KafkaClusterVO, error) {
	version := strings.TrimSpace(req.Auth.Version)
	if version == "" {
		detectedVersion, detectErr := detectKafkaVersion(req.Address, req.Auth, 2500*time.Millisecond)
		if detectErr != nil {
			return nil, fmt.Errorf("Kafka 版本自动探测失败，请手动填写版本后再导入: %w", detectErr)
		}
		version = detectedVersion
	}

	clusterReq := reqKafka.ClusterUpsertRequest{
		Name:               req.Name,
		BootstrapServers:   req.Address,
		Version:            version,
		AuthType:           normalizeDiscoveryAuthType(req.Auth.AuthType),
		Username:           req.Auth.Username,
		Password:           req.Auth.Password,
		TLSEnabled:         req.Auth.TLSEnabled,
		InsecureSkipVerify: req.Auth.InsecureSkipVerify,
		CACert:             req.Auth.CACert,
		ClientCert:         req.Auth.ClientCert,
		ClientKey:          req.Auth.ClientKey,
		Description:        req.Description,
		Environment:        req.Environment,
		Tenant:             req.Tenant,
	}
	return s.CreateCluster(clusterReq)
}

func (s *Service) probeKafkaEndpoint(ip string, port int, timeout time.Duration, auth reqKafka.DiscoveryAuthTemplateRequest) response.KafkaDiscoveryResultVO {
	address := fmt.Sprintf("%s:%d", ip, port)
	result := response.KafkaDiscoveryResultVO{
		IP:      ip,
		Port:    port,
		Address: address,
	}

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		result.ErrorMessage = err.Error()
		return result
	}
	_ = conn.Close()

	version := strings.TrimSpace(auth.Version)
	versionDetectError := ""
	if version == "" {
		detectedVersion, detectErr := detectKafkaVersion(address, auth, timeout)
		if detectErr == nil {
			version = detectedVersion
			result.KafkaVersion = detectedVersion
		} else {
			versionDetectError = detectErr.Error()
			result.VersionDetectError = detectErr.Error()
		}
	}
	if version == "" {
		version = "3.6.0"
	}

	broker := sarama.NewBroker(address)
	auth.Version = version
	config, err := buildDiscoverySaramaConfig(auth, timeout)
	if err != nil {
		result.ErrorMessage = err.Error()
		return result
	}
	if err = broker.Open(config); err != nil && !errors.Is(err, sarama.ErrAlreadyConnected) {
		result.ErrorMessage = err.Error()
		return result
	}
	defer broker.Close()

	req := sarama.NewMetadataRequest(config.Version, nil)
	metaResp, err := broker.GetMetadata(req)
	if err != nil {
		result.ErrorMessage = err.Error()
		return result
	}
	result.LooksLikeKafka = true
	if result.KafkaVersion == "" && versionDetectError == "" {
		result.KafkaVersion = version
	}
	result.BrokerID = broker.ID()
	if metaResp.ControllerID >= 0 {
		result.ControllerID = metaResp.ControllerID
	}
	if metaResp.ClusterID != nil {
		result.ClusterID = *metaResp.ClusterID
	}

	listeners := make([]string, 0, len(metaResp.Brokers))
	for _, item := range metaResp.Brokers {
		if item == nil {
			continue
		}
		addr := item.Addr()
		listeners = append(listeners, addr)
		if item.Addr() == address || strings.HasPrefix(item.Addr(), ip+":") {
			result.BrokerID = item.ID()
		}
	}
	sort.Strings(listeners)
	result.Listeners = listeners
	if versionDetectError != "" && result.ErrorMessage == "" {
		result.ErrorMessage = "Kafka 版本自动探测失败，但已按兼容协议完成识别"
	}
	return result
}

func detectKafkaVersion(address string, auth reqKafka.DiscoveryAuthTemplateRequest, timeout time.Duration) (string, error) {
	candidates := []string{
		"3.9.0",
		"3.8.0",
		"3.7.0",
		"3.6.0",
		"3.5.0",
		"3.4.0",
		"3.3.0",
		"3.2.0",
		"3.1.0",
		"3.0.0",
		"2.8.0",
		"2.7.0",
		"2.6.0",
		"2.5.0",
		"2.4.0",
		"2.3.0",
		"2.2.0",
		"2.1.0",
		"2.0.0",
		"1.1.0",
		"1.0.0",
		"0.11.0.0",
		"0.10.2.0",
		"0.10.1.0",
		"0.10.0.0",
	}
	var lastErr error
	for _, version := range candidates {
		probeAuth := auth
		probeAuth.Version = version
		broker := sarama.NewBroker(address)
		config, err := buildDiscoverySaramaConfig(probeAuth, timeout)
		if err != nil {
			lastErr = err
			continue
		}
		if err = broker.Open(config); err != nil && !errors.Is(err, sarama.ErrAlreadyConnected) {
			lastErr = err
			continue
		}
		req := sarama.NewMetadataRequest(config.Version, nil)
		_, err = broker.GetMetadata(req)
		_ = broker.Close()
		if err == nil {
			return version, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New("未能自动探测 Kafka 版本")
	}
	return "", lastErr
}

func buildDiscoverySaramaConfig(auth reqKafka.DiscoveryAuthTemplateRequest, timeout time.Duration) (*sarama.Config, error) {
	version := strings.TrimSpace(auth.Version)
	if version == "" {
		version = "3.6.0"
	}
	parsedVersion, err := sarama.ParseKafkaVersion(version)
	if err != nil {
		return nil, fmt.Errorf("Kafka 版本解析失败: %w", err)
	}
	config := sarama.NewConfig()
	config.Version = parsedVersion
	config.ClientID = "kafka-console-discovery"
	config.Metadata.Full = true
	config.Metadata.AllowAutoTopicCreation = false
	config.Net.DialTimeout = timeout
	config.Net.ReadTimeout = timeout
	config.Net.WriteTimeout = timeout
	config.Admin.Timeout = timeout

	authType := normalizeDiscoveryAuthType(auth.AuthType)
	if auth.TLSEnabled {
		tlsConfig, err := buildDiscoveryTLSConfig(auth)
		if err != nil {
			return nil, err
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}
	if authType != dal.KafkaAuthTypeNone {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = auth.Username
		config.Net.SASL.Password = auth.Password
		config.Net.SASL.Handshake = true
		config.Net.SASL.Version = sarama.SASLHandshakeV1
		switch authType {
		case dal.KafkaAuthTypePlain:
			config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		case dal.KafkaAuthTypeSCRAMSHA256:
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &XDGSCRAMClient{HashGeneratorFcn: scram.HashGeneratorFcn(sha256.New)}
			}
		case dal.KafkaAuthTypeSCRAMSHA512:
			config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &XDGSCRAMClient{HashGeneratorFcn: scram.HashGeneratorFcn(sha512.New)}
			}
		default:
			return nil, fmt.Errorf("不支持的 Kafka 认证类型: %s", authType)
		}
	}
	return config, config.Validate()
}

func buildDiscoveryTLSConfig(auth reqKafka.DiscoveryAuthTemplateRequest) (*tls.Config, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: auth.InsecureSkipVerify}
	if strings.TrimSpace(auth.CACert) != "" {
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM([]byte(auth.CACert)); !ok {
			return nil, errors.New("Kafka CA 证书解析失败")
		}
		tlsConfig.RootCAs = pool
	}
	if strings.TrimSpace(auth.ClientCert) != "" && strings.TrimSpace(auth.ClientKey) != "" {
		cert, err := tls.X509KeyPair([]byte(auth.ClientCert), []byte(auth.ClientKey))
		if err != nil {
			return nil, errors.New("Kafka 客户端证书解析失败")
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	return tlsConfig, nil
}

func normalizeDiscoveryAuthType(value string) string {
	switch strings.TrimSpace(value) {
	case dal.KafkaAuthTypePlain, dal.KafkaAuthTypeSCRAMSHA256, dal.KafkaAuthTypeSCRAMSHA512:
		return value
	default:
		return dal.KafkaAuthTypeNone
	}
}

func expandCIDR(cidr string) ([]string, error) {
	ip, ipNet, err := net.ParseCIDR(strings.TrimSpace(cidr))
	if err != nil {
		return nil, errors.New("CIDR 格式错误")
	}
	ip = ip.Mask(ipNet.Mask)
	hosts := make([]string, 0)
	for current := dupIP(ip); ipNet.Contains(current); incIP(current) {
		hosts = append(hosts, current.String())
	}
	if len(hosts) <= 2 {
		return hosts, nil
	}
	return hosts[1 : len(hosts)-1], nil
}

func dupIP(ip net.IP) net.IP {
	buf := make(net.IP, len(ip))
	copy(buf, ip)
	return buf
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
