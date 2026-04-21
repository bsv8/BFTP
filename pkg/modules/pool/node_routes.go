package pool

import (
	"context"
	"strings"

	contractprotoid "github.com/bsv8/BFTP-contract/pkg/v1/protoid"
	ncall "github.com/bsv8/BFTP/pkg/infra/ncall"
	poolcore "github.com/bsv8/BFTP/pkg/infra/poolcore"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type NodeRouteRuntime struct {
	BindClientPeer func(clientID string)

	Info       func(clientID string) (poolcore.InfoResp, error)
	Create     func(poolcore.CreateReq) (poolcore.CreateResp, error)
	BaseTx     func(poolcore.BaseTxReq) (poolcore.BaseTxResp, error)
	PayConfirm func(poolcore.PayConfirmReq) (poolcore.PayConfirmResp, error)
	Close      func(poolcore.CloseReq) (poolcore.CloseResp, error)
	State      func(poolcore.StateReq) (poolcore.StateResp, error)
}

// HandlePoolProto 统一挂载 pool 独立协议。
// 硬切说明：不再使用 node.call + route 分发模型，改为独立 protocol.ID 直连。
func HandlePoolProto(rt NodeRouteRuntime) []protocolRegistration {
	return []protocolRegistration{
		{Proto: contractprotoid.ProtoPoolV1Info, Req: poolcore.InfoReq{}, F: handlePoolInfo},
		{Proto: contractprotoid.ProtoPoolV1Create, Req: poolcore.CreateReq{}, F: handlePoolCreate},
		{Proto: contractprotoid.ProtoPoolV1BaseTx, Req: poolcore.BaseTxReq{}, F: handlePoolBaseTx},
		{Proto: contractprotoid.ProtoPoolV1PayConfirm, Req: poolcore.PayConfirmReq{}, F: handlePoolPayConfirm},
		{Proto: contractprotoid.ProtoPoolV1Close, Req: poolcore.CloseReq{}, F: handlePoolClose},
		{Proto: contractprotoid.ProtoPoolV1SessionState, Req: poolcore.StateReq{}, F: handlePoolSessionState},
	}
}

type protocolRegistration struct {
	Proto protocol.ID
	Req   interface{}
	F     func(context.Context, NodeRouteRuntime, ncall.CallContext, interface{}) (ncall.CallResp, error)
}

func handlePoolInfo(ctx context.Context, rt NodeRouteRuntime, meta ncall.CallContext, body interface{}) (ncall.CallResp, error) {
	_ = body
	clientID := strings.TrimSpace(meta.SenderPubkeyHex)
	if rt.BindClientPeer != nil {
		rt.BindClientPeer(clientID)
	}
	if rt.Info == nil {
		return ncall.CallResp{Ok: false, Code: "ROUTE_NOT_FOUND", Message: "route not found"}, nil
	}
	resp, err := rt.Info(clientID)
	if err != nil {
		return ncall.CallResp{}, err
	}
	return ncall.MarshalProto(&resp)
}

func handlePoolCreate(ctx context.Context, rt NodeRouteRuntime, meta ncall.CallContext, body interface{}) (ncall.CallResp, error) {
	req := body.(poolcore.CreateReq)
	req.ClientID = strings.TrimSpace(meta.SenderPubkeyHex)
	if rt.BindClientPeer != nil {
		rt.BindClientPeer(req.ClientID)
	}
	if rt.Create == nil {
		return ncall.CallResp{Ok: false, Code: "ROUTE_NOT_FOUND", Message: "route not found"}, nil
	}
	resp, err := rt.Create(req)
	if err != nil {
		return ncall.CallResp{}, err
	}
	return ncall.MarshalProto(&resp)
}

func handlePoolBaseTx(ctx context.Context, rt NodeRouteRuntime, meta ncall.CallContext, body interface{}) (ncall.CallResp, error) {
	req := body.(poolcore.BaseTxReq)
	req.ClientID = strings.TrimSpace(meta.SenderPubkeyHex)
	if rt.BaseTx == nil {
		return ncall.CallResp{Ok: false, Code: "ROUTE_NOT_FOUND", Message: "route not found"}, nil
	}
	resp, err := rt.BaseTx(req)
	if err != nil {
		return ncall.CallResp{}, err
	}
	return ncall.MarshalProto(&resp)
}

func handlePoolPayConfirm(ctx context.Context, rt NodeRouteRuntime, meta ncall.CallContext, body interface{}) (ncall.CallResp, error) {
	req := body.(poolcore.PayConfirmReq)
	req.ClientID = strings.TrimSpace(meta.SenderPubkeyHex)
	if rt.PayConfirm == nil {
		return ncall.CallResp{Ok: false, Code: "ROUTE_NOT_FOUND", Message: "route not found"}, nil
	}
	resp, err := rt.PayConfirm(req)
	if err != nil {
		return ncall.CallResp{}, err
	}
	return ncall.MarshalProto(&resp)
}

func handlePoolClose(ctx context.Context, rt NodeRouteRuntime, meta ncall.CallContext, body interface{}) (ncall.CallResp, error) {
	req := body.(poolcore.CloseReq)
	req.ClientID = strings.TrimSpace(meta.SenderPubkeyHex)
	if rt.Close == nil {
		return ncall.CallResp{Ok: false, Code: "ROUTE_NOT_FOUND", Message: "route not found"}, nil
	}
	resp, err := rt.Close(req)
	if err != nil {
		return ncall.CallResp{}, err
	}
	return ncall.MarshalProto(&resp)
}

func handlePoolSessionState(ctx context.Context, rt NodeRouteRuntime, meta ncall.CallContext, body interface{}) (ncall.CallResp, error) {
	req := body.(poolcore.StateReq)
	req.ClientID = strings.TrimSpace(meta.SenderPubkeyHex)
	if rt.State == nil {
		return ncall.CallResp{Ok: false, Code: "ROUTE_NOT_FOUND", Message: "route not found"}, nil
	}
	resp, err := rt.State(req)
	if err != nil {
		return ncall.CallResp{}, err
	}
	return ncall.MarshalProto(&resp)
}

// HandleNodeCall 统一挂载 pool.v1.* 的 node.call 路由。
// 硬切说明：已废弃，请使用 HandlePoolProto 注册独立协议。
func HandleNodeCall(_ context.Context, rt NodeRouteRuntime, meta ncall.CallContext, req ncall.CallReq) (bool, ncall.CallResp, error) {
	switch strings.TrimSpace(req.Route) {
	case ncall.RoutePoolV1Info:
		var body poolcore.InfoReq
		if err := ncall.DecodeProto(req.Route, req.Body, &body, false); err != nil {
			return true, ncall.CallResp{Ok: false, Code: "BAD_REQUEST", Message: err.Error()}, nil
		}
		clientID := strings.TrimSpace(meta.SenderPubkeyHex)
		if rt.BindClientPeer != nil {
			rt.BindClientPeer(clientID)
		}
		if rt.Info == nil {
			return true, ncall.CallResp{Ok: false, Code: "ROUTE_NOT_FOUND", Message: "route not found"}, nil
		}
		resp, err := rt.Info(clientID)
		if err != nil {
			return true, ncall.CallResp{}, err
		}
		out, err := ncall.MarshalProto(&resp)
		return true, out, err
	case ncall.RoutePoolV1Create:
		var body poolcore.CreateReq
		if err := ncall.DecodeProto(req.Route, req.Body, &body, true); err != nil {
			return true, ncall.CallResp{Ok: false, Code: "BAD_REQUEST", Message: err.Error()}, nil
		}
		body.ClientID = strings.TrimSpace(meta.SenderPubkeyHex)
		if rt.BindClientPeer != nil {
			rt.BindClientPeer(body.ClientID)
		}
		if rt.Create == nil {
			return true, ncall.CallResp{Ok: false, Code: "ROUTE_NOT_FOUND", Message: "route not found"}, nil
		}
		resp, err := rt.Create(body)
		if err != nil {
			return true, ncall.CallResp{}, err
		}
		out, err := ncall.MarshalProto(&resp)
		return true, out, err
	case ncall.RoutePoolV1BaseTx:
		var body poolcore.BaseTxReq
		if err := ncall.DecodeProto(req.Route, req.Body, &body, true); err != nil {
			return true, ncall.CallResp{Ok: false, Code: "BAD_REQUEST", Message: err.Error()}, nil
		}
		body.ClientID = strings.TrimSpace(meta.SenderPubkeyHex)
		if rt.BaseTx == nil {
			return true, ncall.CallResp{Ok: false, Code: "ROUTE_NOT_FOUND", Message: "route not found"}, nil
		}
		resp, err := rt.BaseTx(body)
		if err != nil {
			return true, ncall.CallResp{}, err
		}
		out, err := ncall.MarshalProto(&resp)
		return true, out, err
	case ncall.RoutePoolV1PayConfirm:
		var body poolcore.PayConfirmReq
		if err := ncall.DecodeProto(req.Route, req.Body, &body, true); err != nil {
			return true, ncall.CallResp{Ok: false, Code: "BAD_REQUEST", Message: err.Error()}, nil
		}
		body.ClientID = strings.TrimSpace(meta.SenderPubkeyHex)
		if rt.PayConfirm == nil {
			return true, ncall.CallResp{Ok: false, Code: "ROUTE_NOT_FOUND", Message: "route not found"}, nil
		}
		resp, err := rt.PayConfirm(body)
		if err != nil {
			return true, ncall.CallResp{}, err
		}
		out, err := ncall.MarshalProto(&resp)
		return true, out, err
	case ncall.RoutePoolV1Close:
		var body poolcore.CloseReq
		if err := ncall.DecodeProto(req.Route, req.Body, &body, true); err != nil {
			return true, ncall.CallResp{Ok: false, Code: "BAD_REQUEST", Message: err.Error()}, nil
		}
		body.ClientID = strings.TrimSpace(meta.SenderPubkeyHex)
		if rt.Close == nil {
			return true, ncall.CallResp{Ok: false, Code: "ROUTE_NOT_FOUND", Message: "route not found"}, nil
		}
		resp, err := rt.Close(body)
		if err != nil {
			return true, ncall.CallResp{}, err
		}
		out, err := ncall.MarshalProto(&resp)
		return true, out, err
	case ncall.RoutePoolV1SessionState:
		var body poolcore.StateReq
		if err := ncall.DecodeProto(req.Route, req.Body, &body, false); err != nil {
			return true, ncall.CallResp{Ok: false, Code: "BAD_REQUEST", Message: err.Error()}, nil
		}
		body.ClientID = strings.TrimSpace(meta.SenderPubkeyHex)
		if rt.State == nil {
			return true, ncall.CallResp{Ok: false, Code: "ROUTE_NOT_FOUND", Message: "route not found"}, nil
		}
		resp, err := rt.State(body)
		if err != nil {
			return true, ncall.CallResp{}, err
		}
		out, err := ncall.MarshalProto(&resp)
		return true, out, err
	default:
		return false, ncall.CallResp{}, nil
	}
}
