package pool

import (
	"context"
	"strings"

	ncall "github.com/bsv8/BFTP/pkg/infra/ncall"
	poolcore "github.com/bsv8/BFTP/pkg/infra/poolcore"
)

type NodeRouteRuntime struct {
	BindClientPeer func(clientID string)

	Info         func(clientID string) (poolcore.InfoResp, error)
	Create       func(poolcore.CreateReq) (poolcore.CreateResp, error)
	BaseTx       func(poolcore.BaseTxReq) (poolcore.BaseTxResp, error)
	PayConfirm   func(poolcore.PayConfirmReq) (poolcore.PayConfirmResp, error)
	Close        func(poolcore.CloseReq) (poolcore.CloseResp, error)
	State        func(poolcore.StateReq) (poolcore.StateResp, error)
}

// HandleNodeCall 统一挂载 pool.v1.* 的 node.call 路由。
// 设计说明：
// - pool 基础路由是公开能力的共用壳，不应该在 gateway/domain/arbiter 各写一遍；
// - 业务侧只注入“具体怎么报价/创建/扣费/收尾”的函数；
// - 这样 pool 模块和角色实现之间的边界会更清楚。
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
