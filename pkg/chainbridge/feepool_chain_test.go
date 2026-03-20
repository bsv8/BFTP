package chainbridge

import "testing"

func TestResolvePortBaseURLFromEnv(t *testing.T) {
	t.Setenv(FeePoolChainBaseURLEnv, " http://127.0.0.1:18222/ ")
	if got, want := ResolvePortBaseURLFromEnv(), "http://127.0.0.1:18222"; got != want {
		t.Fatalf("ResolvePortBaseURLFromEnv()=%q, want %q", got, want)
	}
}

func TestNewDefaultFeePoolChainUsesEnvPort(t *testing.T) {
	t.Setenv(FeePoolChainBaseURLEnv, "http://127.0.0.1:18299/")
	chain, err := NewDefaultFeePoolChain(RouteConfig{Network: "test"}, 0)
	if err != nil {
		t.Fatalf("NewDefaultFeePoolChain() error: %v", err)
	}
	if got, want := chain.BaseURL(), "http://127.0.0.1:18299"; got != want {
		t.Fatalf("BaseURL()=%q, want %q", got, want)
	}
}
