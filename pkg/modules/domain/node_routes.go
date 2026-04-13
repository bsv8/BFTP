package domainmodule

import (
	"context"
	"strings"

	bfterrors "github.com/bsv8/BFTP-contract/pkg/v1/errors"
	ncall "github.com/bsv8/BFTP/pkg/infra/ncall"
)

// NodeRouteRuntime 描述 domain.v1.* 公开 node.call 路由在角色侧需要注入的运行时。
// 设计说明：
// - modules.domain 只收口公开合同壳、解码和基础入参校验；
// - 具体业务处理、费用池扣费、持久化仍留在角色项目；
// - 这样后续 domain/gateway/client 若复用同一合同面，边界会更清楚。
type NodeRouteRuntime struct {
	Pricing        func(context.Context) (DomainPricingBody, error)
	Resolve        func(context.Context, ncall.CallContext, ncall.CallReq, NameRouteReq) (ncall.CallResp, error)
	Query          func(context.Context, ncall.CallContext, ncall.CallReq, NameRouteReq) (ncall.CallResp, error)
	Lock           func(context.Context, ncall.CallContext, ncall.CallReq, NameTargetRouteReq) (ncall.CallResp, error)
	ListOwned      func(context.Context, ncall.CallContext, ListOwnedReq) (ListOwnedResp, error)
	SetTarget      func(context.Context, ncall.CallContext, ncall.CallReq, NameTargetRouteReq) (ncall.CallResp, error)
	RegisterSubmit func(context.Context, ncall.CallContext, RegisterSubmitReq) (RegisterSubmitResp, error)
}

// HandleNodeCall 统一处理 domain.v1.* 的 node.call 路由。
// 注意：
// - 它只做共享分发，不直接调用 ncall.Register；
// - 真正把入口挂到宿主上的动作，仍由角色项目在自己的 run/register 阶段显式完成。
func HandleNodeCall(ctx context.Context, rt NodeRouteRuntime, meta ncall.CallContext, req ncall.CallReq) (bool, ncall.CallResp, error) {
	switch strings.TrimSpace(req.Route) {
	case RouteDomainV1Pricing:
		if rt.Pricing == nil {
			return true, routeNotFoundResp(), nil
		}
		resp, err := rt.Pricing(ctx)
		if err != nil {
			return true, ncall.CallResp{}, err
		}
		out, err := ncall.MarshalProto(&resp)
		return true, out, err
	case RouteDomainV1Resolve:
		var body NameRouteReq
		if err := ncall.DecodeProto(req.Route, req.Body, &body, true); err != nil {
			return true, badRequestResp(err.Error()), nil
		}
		if rt.Resolve == nil {
			return true, routeNotFoundResp(), nil
		}
		resp, err := rt.Resolve(ctx, meta, req, body)
		return true, resp, err
	case RouteDomainV1Query:
		var body NameRouteReq
		if err := ncall.DecodeProto(req.Route, req.Body, &body, true); err != nil {
			return true, badRequestResp(err.Error()), nil
		}
		if rt.Query == nil {
			return true, routeNotFoundResp(), nil
		}
		resp, err := rt.Query(ctx, meta, req, body)
		return true, resp, err
	case RouteDomainV1Lock:
		var body NameTargetRouteReq
		if err := ncall.DecodeProto(req.Route, req.Body, &body, true); err != nil {
			return true, badRequestResp(err.Error()), nil
		}
		if rt.Lock == nil {
			return true, routeNotFoundResp(), nil
		}
		resp, err := rt.Lock(ctx, meta, req, body)
		return true, resp, err
	case RouteDomainV1ListOwned:
		var body ListOwnedReq
		if err := ncall.DecodeProto(req.Route, req.Body, &body, false); err != nil {
			return true, badRequestResp(err.Error()), nil
		}
		if strings.TrimSpace(body.OwnerPubkeyHex) == "" {
			body.OwnerPubkeyHex = strings.TrimSpace(meta.SenderPubkeyHex)
		}
		if rt.ListOwned == nil {
			return true, routeNotFoundResp(), nil
		}
		resp, err := rt.ListOwned(ctx, meta, body)
		if err != nil {
			return true, ncall.CallResp{}, err
		}
		out, err := ncall.MarshalProto(&resp)
		return true, out, err
	case RouteDomainV1SetTarget:
		var body NameTargetRouteReq
		if err := ncall.DecodeProto(req.Route, req.Body, &body, true); err != nil {
			return true, badRequestResp(err.Error()), nil
		}
		if rt.SetTarget == nil {
			return true, routeNotFoundResp(), nil
		}
		resp, err := rt.SetTarget(ctx, meta, req, body)
		return true, resp, err
	case RouteDomainV1RegisterSubmit:
		var body RegisterSubmitReq
		if err := ncall.DecodeProto(req.Route, req.Body, &body, true); err != nil {
			return true, badRequestResp(err.Error()), nil
		}
		if len(body.RegisterTx) == 0 {
			return true, badRequestResp("register_tx required"), nil
		}
		body.ClientID = strings.TrimSpace(meta.SenderPubkeyHex)
		body.RegisterTx = append([]byte(nil), body.RegisterTx...)
		if rt.RegisterSubmit == nil {
			return true, routeNotFoundResp(), nil
		}
		resp, err := rt.RegisterSubmit(ctx, meta, body)
		if err != nil {
			return true, ncall.CallResp{}, err
		}
		out, err := ncall.MarshalProto(&resp)
		return true, out, err
	default:
		return false, ncall.CallResp{}, nil
	}
}

func badRequestResp(message string) ncall.CallResp {
	return ncall.CallResp{Ok: false, Code: string(bfterrors.CodeBadRequest), Message: message}
}

func routeNotFoundResp() ncall.CallResp {
	return ncall.CallResp{Ok: false, Code: string(bfterrors.CodeRouteNotFound), Message: "route not found"}
}
