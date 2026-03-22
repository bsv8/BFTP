package chainbridge

import (
	"testing"

	chainapi "github.com/bsv8/BSVChainAPI"
)

func TestResolvePortBaseURLFromEnv(t *testing.T) {
	t.Setenv(FeePoolChainBaseURLEnv, " http://127.0.0.1:18222/ ")
	if got, want := ResolvePortBaseURLFromEnv(), "http://127.0.0.1:18222"; got != want {
		t.Fatalf("ResolvePortBaseURLFromEnv()=%q, want %q", got, want)
	}
}

func TestNewDefaultFeePoolChainUsesEnvPort(t *testing.T) {
	t.Setenv(FeePoolChainBaseURLEnv, "http://127.0.0.1:18299/")
	chain, err := NewDefaultFeePoolChain(RouteConfig{Provider: chainapi.WhatsOnChainProvider, Network: "test"}, 0)
	if err != nil {
		t.Fatalf("NewDefaultFeePoolChain() error: %v", err)
	}
	if got, want := chain.BaseURL(), "http://127.0.0.1:18299"; got != want {
		t.Fatalf("BaseURL()=%q, want %q", got, want)
	}
	if got, want := chain.ReadRoute(), (chainapi.Route{Provider: chainapi.WhatsOnChainProvider, Network: "test", Profile: "default"}); got != want {
		t.Fatalf("ReadRoute()=%+v, want %+v", got, want)
	}
	if got := chain.TxSubmitPolicy(); len(got.Routes) != 1 || got.Routes[0] != chain.ReadRoute() {
		t.Fatalf("unexpected TxSubmitPolicy(): %+v", got)
	}
}

func TestNewEmbeddedFeePoolChainRequiresExplicitProvider(t *testing.T) {
	if _, err := NewEmbeddedFeePoolChain(RouteConfig{Network: "test"}, 0); err == nil {
		t.Fatalf("expected explicit provider error")
	}
}

func TestNewEmbeddedFeePoolChainRejectsBroadcastOnlyReadRoute(t *testing.T) {
	_, err := NewEmbeddedFeePoolChainFromPlan(FeePoolChainPlan{
		Routes: []chainapi.RouteConfig{
			{Provider: chainapi.GorillaPoolARCProvider, Network: "main"},
		},
		ReadRoute: chainapi.Route{Provider: chainapi.GorillaPoolARCProvider, Network: "main"},
		TxSubmitPolicy: chainapi.TxSubmitPolicy{
			Routes: []chainapi.Route{
				{Provider: chainapi.GorillaPoolARCProvider, Network: "main"},
			},
		},
	})
	if err == nil {
		t.Fatalf("expected read route capability error")
	}
}
