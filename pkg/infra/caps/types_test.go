package caps

import (
	"context"
	"strings"
	"testing"

	ncall "github.com/bsv8/BFTP/pkg/infra/ncall"
)

func TestAssembleBuildsBundleAndShowBody(t *testing.T) {
	bundle, err := Assemble(
		ModuleSpec{
			InternalAbility: "bftp.capabilities@1",
			Routes:          []string{"node.v1.capabilities_show"},
		},
		ModuleSpec{
			InternalAbility: "bftp.pool@1",
			PublicCapability: &PublicCapability{
				ID:      "pool",
				Version: 1,
			},
			Routes: []string{"pool.v1.info"},
		},
		ModuleSpec{
			InternalAbility: "bftp.broadcast@1",
			PublicCapability: &PublicCapability{
				ID:      "broadcast",
				Version: 1,
			},
			Protos:    []string{"/bsv-transfer/demand/publish_paid/1.0.0"},
			HTTPPaths: []string{"/api/v1/admin/clients"},
		},
	)
	if err != nil {
		t.Fatalf("Assemble() failed: %v", err)
	}
	if len(bundle.Modules) != 3 {
		t.Fatalf("unexpected module count: %d", len(bundle.Modules))
	}
	if len(bundle.InternalAbilities) != 3 {
		t.Fatalf("unexpected internal ability count: %d", len(bundle.InternalAbilities))
	}
	if len(bundle.PublicCapabilities) != 2 {
		t.Fatalf("unexpected public capability count: %d", len(bundle.PublicCapabilities))
	}
	if len(bundle.Routes) != 2 {
		t.Fatalf("unexpected route count: %d", len(bundle.Routes))
	}
	if len(bundle.Protos) != 1 {
		t.Fatalf("unexpected proto count: %d", len(bundle.Protos))
	}
	if len(bundle.HTTPPaths) != 1 {
		t.Fatalf("unexpected http path count: %d", len(bundle.HTTPPaths))
	}

	body := bundle.ShowBody(" 02AA ")
	if body.NodePubkeyHex != "02aa" {
		t.Fatalf("unexpected node pubkey hex: %q", body.NodePubkeyHex)
	}
	if len(body.Capabilities) != 2 {
		t.Fatalf("unexpected capability item count: %d", len(body.Capabilities))
	}
	if body.Capabilities[0].ID != "pool" || body.Capabilities[0].Version != 1 {
		t.Fatalf("unexpected first capability: %+v", body.Capabilities[0])
	}
	if body.Capabilities[1].ID != "broadcast" || body.Capabilities[1].Version != 1 {
		t.Fatalf("unexpected second capability: %+v", body.Capabilities[1])
	}
}

func TestAssembleRejectsDuplicateRoute(t *testing.T) {
	_, err := Assemble(
		ModuleSpec{InternalAbility: "bftp.first@1", Routes: []string{"node.v1.capabilities_show"}},
		ModuleSpec{InternalAbility: "bftp.second@1", Routes: []string{"node.v1.capabilities_show"}},
	)
	if err == nil {
		t.Fatal("expected route conflict error")
	}
	if !strings.Contains(err.Error(), "route conflict") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAssembleRejectsDuplicateProto(t *testing.T) {
	_, err := Assemble(
		ModuleSpec{InternalAbility: "bftp.first@1", Protos: []string{"/same/proto/1.0.0"}},
		ModuleSpec{InternalAbility: "bftp.second@1", Protos: []string{"/same/proto/1.0.0"}},
	)
	if err == nil {
		t.Fatal("expected proto conflict error")
	}
	if !strings.Contains(err.Error(), "proto conflict") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAssembleRejectsDuplicatePublicCapability(t *testing.T) {
	_, err := Assemble(
		ModuleSpec{
			InternalAbility: "bftp.first@1",
			PublicCapability: &PublicCapability{
				ID:      "pool",
				Version: 1,
			},
		},
		ModuleSpec{
			InternalAbility: "bftp.second@1",
			PublicCapability: &PublicCapability{
				ID:      "pool",
				Version: 1,
			},
		},
	)
	if err == nil {
		t.Fatal("expected public capability conflict error")
	}
	if !strings.Contains(err.Error(), "public capability conflict") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAssembleRejectsDuplicateHTTPPath(t *testing.T) {
	_, err := Assemble(
		ModuleSpec{InternalAbility: "bftp.first@1", HTTPPaths: []string{"/api/v1/info"}},
		ModuleSpec{InternalAbility: "bftp.second@1", HTTPPaths: []string{"/api/v1/info"}},
	)
	if err == nil {
		t.Fatal("expected http path conflict error")
	}
	if !strings.Contains(err.Error(), "http path conflict") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAssembleRejectsDuplicateHTTPPathInsideSingleSpec(t *testing.T) {
	_, err := Assemble(
		ModuleSpec{InternalAbility: "bftp.http@1", HTTPPaths: []string{"/api/v1/info", "/api/v1/info"}},
	)
	if err == nil {
		t.Fatal("expected duplicate http path error")
	}
	if !strings.Contains(err.Error(), "http path duplicate") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChainNodeCallHandlesCapabilitiesAndSegments(t *testing.T) {
	handler := ChainNodeCall(
		func() ncall.CapabilitiesShowBody {
			return ncall.CapabilitiesShowBody{
				NodePubkeyHex: "02aa",
				Capabilities: []*ncall.CapabilityItem{
					{ID: "pool", Version: 1},
				},
			}
		},
		func(_ context.Context, _ ncall.CallContext, req ncall.CallReq) (bool, ncall.CallResp, error) {
			if req.Route != "pool.v1.info" {
				return false, ncall.CallResp{}, nil
			}
			return true, ncall.CallResp{Ok: true, Code: "OK", Message: "handled"}, nil
		},
	)

	showResp, err := handler(context.Background(), ncall.CallContext{}, ncall.CallReq{Route: ncall.RouteNodeV1CapabilitiesShow})
	if err != nil {
		t.Fatalf("capabilities_show failed: %v", err)
	}
	if !showResp.Ok {
		t.Fatalf("capabilities_show not ok: %+v", showResp)
	}
	segResp, err := handler(context.Background(), ncall.CallContext{}, ncall.CallReq{Route: "pool.v1.info"})
	if err != nil {
		t.Fatalf("segment route failed: %v", err)
	}
	if !segResp.Ok || segResp.Code != "OK" {
		t.Fatalf("unexpected segment response: %+v", segResp)
	}
	missResp, err := handler(context.Background(), ncall.CallContext{}, ncall.CallReq{Route: "unknown.route"})
	if err != nil {
		t.Fatalf("unknown route failed: %v", err)
	}
	if missResp.Ok || missResp.Code != "ROUTE_NOT_FOUND" {
		t.Fatalf("unexpected miss response: %+v", missResp)
	}
}
