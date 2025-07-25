// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package features

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestFeatureSetMatchRequirements(t *testing.T) {
	features := Set{}
	matches, _ := features.MatchRequirements()
	if !matches {
		t.Error("empty requirements should always match")
	}
	matches, _ = features.MatchRequirements(RequireEnabled(L7Proxy))
	if matches {
		t.Error("empty features should not match any requirement")
	}

	features[L7Proxy] = Status{
		Enabled: true,
	}
	matches, _ = features.MatchRequirements()
	if !matches {
		t.Error("empty requirements should always match")
	}
	matches, _ = features.MatchRequirements(RequireEnabled(L7Proxy))
	if !matches {
		t.Errorf("expected features %v to match feature %v", features, L7Proxy)
	}

	cniMode := "aws-cni"
	features[CNIChaining] = Status{
		Enabled: true,
		Mode:    cniMode,
	}
	matches, _ = features.MatchRequirements()
	if !matches {
		t.Error("empty requirements should always match")
	}
	matches, _ = features.MatchRequirements(RequireEnabled(L7Proxy))
	if !matches {
		t.Errorf("expected features %v to match feature %v", features, L7Proxy)
	}
	matches, _ = features.MatchRequirements(RequireEnabled(CNIChaining), RequireMode(CNIChaining, cniMode))
	if !matches {
		t.Errorf("expected features %v to match feature %v with mode %v", features, CNIChaining, cniMode)
	}
	cniMode = "generic-veth"
	matches, _ = features.MatchRequirements(RequireEnabled(CNIChaining), RequireMode(CNIChaining, cniMode))
	if matches {
		t.Errorf("features %v unexpectedly matched feature %v with mode %v", features, CNIChaining, cniMode)
	}

	matches, _ = features.MatchRequirements(RequireEnabled(CNIChaining), RequireModeIsNot(CNIChaining, "kubernetes"))
	if !matches {
		t.Errorf("expected features %v to match feature %v with mode different from kubernetes", features, CNIChaining)
	}

	matches, _ = features.MatchRequirements(RequireEnabled(CNIChaining), RequireModeIsNot(CNIChaining, "aws-cni"))
	if matches {
		t.Errorf("expected features %v to not match feature %v with mode different from aws-cni", features, CNIChaining)
	}

}

func TestFeatureSet_extractTunnelFromConfigMap(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string

		enabled bool
		proto   string
	}{
		{
			name:    "empty config map",
			enabled: true,
			proto:   "vxlan",
		},
		{
			name:    "native routing, default protocol",
			data:    map[string]string{"routing-mode": "native"},
			enabled: false,
			proto:   "vxlan",
		},
		{
			name:    "native routing, geneve protocol",
			data:    map[string]string{"routing-mode": "native", "tunnel-protocol": "geneve"},
			enabled: false,
			proto:   "geneve",
		},
		{
			name:    "tunnel routing, default protocol",
			data:    map[string]string{"routing-mode": "tunnel"},
			enabled: true,
			proto:   "vxlan",
		},
		{
			name:    "tunnel routing, geneve protocol",
			data:    map[string]string{"routing-mode": "tunnel", "tunnel-protocol": "geneve"},
			enabled: true,
			proto:   "geneve",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := Set{}
			cm := corev1.ConfigMap{Data: tt.data}
			fs.ExtractFromConfigMap(&cm)

			assert.Equal(t, tt.enabled, fs[Tunnel].Enabled)
			assert.Equal(t, tt.proto, fs[Tunnel].Mode)
		})
	}
}

func TestFeatureSet_extractFromConfigMap(t *testing.T) {
	fs := Set{}
	cm := corev1.ConfigMap{}
	fs.ExtractFromConfigMap(&cm)
	cm.Data = map[string]string{
		"enable-ipv4":                  "true",
		"enable-ipv6":                  "true",
		"mesh-auth-mutual-enabled":     "true",
		"enable-egress-gateway":        "true",
		"ipam":                         "eni",
		"enable-ipsec":                 "true",
		"enable-local-redirect-policy": "true",
		"bpf-lb-external-clusterip":    "true",
		"enable-bgp-control-plane":     "true",
	}
	fs.ExtractFromConfigMap(&cm)
	assert.True(t, fs[IPv4].Enabled)
	assert.True(t, fs[IPv6].Enabled)
	assert.True(t, fs[AuthSpiffe].Enabled)
	assert.True(t, fs[EgressGateway].Enabled)
	assert.True(t, fs[IPsecEnabled].Enabled)
	assert.True(t, fs[LocalRedirectPolicy].Enabled)
	assert.True(t, fs[BPFLBExternalClusterIP].Enabled)
	assert.True(t, fs[BGPControlPlane].Enabled)
	assert.Equal(t, "eni", fs[CiliumIPAMMode].Mode)
}
