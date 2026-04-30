package kafka

import (
	"strings"
	"testing"
	"time"

	"devops-console-backend/internal/dal"
	reqKafka "devops-console-backend/internal/dal/request/kafka"
)

func TestBuildClusterModelRejectsChangedClientCertWithoutKey(t *testing.T) {
	existing := &dal.KafkaCluster{
		Name:                "cluster-a",
		BootstrapServers:    "broker-a:9092",
		Version:             "3.6.0",
		ClientCert:          "old-cert",
		ClientKeyCiphertext: "encrypted-old-key",
	}

	_, err := buildClusterModel(existing, reqKafka.ClusterUpsertRequest{
		Name:             "cluster-a",
		BootstrapServers: "broker-a:9092",
		Version:          "3.6.0",
		ClientCert:       "new-cert",
	})
	if err == nil || !strings.Contains(err.Error(), "客户端私钥") {
		t.Fatalf("expected missing client key error, got %v", err)
	}
}

func TestBuildClusterModelKeepsExistingClientKeyWhenClientCertUnchanged(t *testing.T) {
	existing := &dal.KafkaCluster{
		Name:                "cluster-a",
		BootstrapServers:    "broker-a:9092",
		Version:             "3.6.0",
		ClientCert:          "same-cert",
		ClientKeyCiphertext: "encrypted-old-key",
	}

	cluster, err := buildClusterModel(existing, reqKafka.ClusterUpsertRequest{
		Name:             "cluster-a",
		BootstrapServers: "broker-a:9092",
		Version:          "3.6.0",
		ClientCert:       "same-cert",
	})
	if err != nil {
		t.Fatalf("expected unchanged cert to keep old key, got %v", err)
	}
	if cluster.ClientKeyCiphertext != "encrypted-old-key" {
		t.Fatalf("expected old client key ciphertext to be preserved, got %q", cluster.ClientKeyCiphertext)
	}
}

func TestExpandCIDRRejectsLargeIPv4Range(t *testing.T) {
	_, err := expandCIDR("10.0.0.0/16", 128)
	if err == nil || !strings.Contains(err.Error(), "最多允许扫描 128 个主机") {
		t.Fatalf("expected large CIDR error, got %v", err)
	}
}

func TestExpandCIDRAcceptsSmallIPv4Range(t *testing.T) {
	hosts, err := expandCIDR("10.0.0.0/30", 16)
	if err != nil {
		t.Fatalf("expected /30 CIDR to pass, got %v", err)
	}
	if len(hosts) != 2 || hosts[0] != "10.0.0.1" || hosts[1] != "10.0.0.2" {
		t.Fatalf("unexpected hosts: %#v", hosts)
	}
}

func TestMessageBrowseTimeoutScalesWithLimit(t *testing.T) {
	if got := messageBrowseTimeout(20); got != 3*time.Second {
		t.Fatalf("expected base timeout for small limit, got %s", got)
	}
	if got := messageBrowseTimeout(120); got != 5*time.Second {
		t.Fatalf("expected scaled timeout for medium limit, got %s", got)
	}
	if got := messageBrowseTimeout(500); got != 12*time.Second {
		t.Fatalf("expected capped timeout for large limit, got %s", got)
	}
}

func TestBuildConsumerGroupLagWarning(t *testing.T) {
	if got := buildConsumerGroupLagWarning(0, 0); got != "" {
		t.Fatalf("expected empty warning, got %q", got)
	}
	if got := buildConsumerGroupLagWarning(2, 0); !strings.Contains(got, "2 个分区 offset 查询失败") {
		t.Fatalf("expected offset warning, got %q", got)
	}
	if got := buildConsumerGroupLagWarning(1, 3); !strings.Contains(got, "1 个分区 offset 查询失败") || !strings.Contains(got, "3 个分区最新 offset 查询失败") {
		t.Fatalf("expected combined warning, got %q", got)
	}
}

func TestTruncatePreviewTextKeepsUTF8Integrity(t *testing.T) {
	input := "你好Kafka世界"
	got := truncatePreviewText(input, 3)
	if got != "你好K..." {
		t.Fatalf("unexpected truncated preview: %q", got)
	}
}

func TestVersionDetectBudgetUsesReasonableBounds(t *testing.T) {
	if got := versionDetectBudget(500 * time.Millisecond); got != 4*time.Second {
		t.Fatalf("expected minimum budget, got %s", got)
	}
	if got := versionDetectBudget(3 * time.Second); got != 10*time.Second {
		t.Fatalf("expected capped budget, got %s", got)
	}
	if got := versionDetectBudget(2 * time.Second); got != 8*time.Second {
		t.Fatalf("expected scaled budget, got %s", got)
	}
}

func TestToClusterVOHidesCertificateBodies(t *testing.T) {
	cluster := dal.KafkaCluster{
		ID:                  1,
		Name:                "cluster-a",
		BootstrapServers:    "broker-a:9092",
		Version:             "3.6.0",
		CACert:              "ca-body",
		ClientCert:          "client-cert-body",
		ClientKeyCiphertext: "ciphertext",
	}

	vo := toClusterVO(cluster)
	if !vo.HasCACert || !vo.HasClientCert || !vo.HasClientKey {
		t.Fatalf("expected certificate presence flags to be preserved: %#v", vo)
	}
}

func TestToClusterDetailVORetainsCertificateBodies(t *testing.T) {
	cluster := dal.KafkaCluster{
		ID:         1,
		Name:       "cluster-a",
		CACert:     "ca-body",
		ClientCert: "client-cert-body",
	}

	detail := toClusterDetailVO(cluster)
	if detail.CACert != "ca-body" || detail.ClientCert != "client-cert-body" {
		t.Fatalf("expected detail response to include certificate bodies: %#v", detail)
	}
}
