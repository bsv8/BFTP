package caps

import (
	"context"
	"net/http"
	"strings"

	bfterrors "github.com/bsv8/BFTP-contract/pkg/v1/errors"
	ncall "github.com/bsv8/BFTP/pkg/infra/ncall"
	"github.com/bsv8/BFTP/pkg/infra/pproto"
	"github.com/libp2p/go-libp2p/core/host"
	coreprotocol "github.com/libp2p/go-libp2p/core/protocol"
)

// MountContext 是当前这套静态装配层的最小挂载上下文。
// 设计说明：
// - 这里只统一宿主、默认安全域和本地 HTTP mux；
// - 不把角色私有运行时、业务依赖、生命周期控制揉进来；
// - 目标是减少重复挂载代码，而不是引入重型插件系统。
type MountContext struct {
	Host          host.Host
	NodeSecurity  pproto.SecurityConfig
	ProtoSecurity pproto.SecurityConfig
	AdminMux      *http.ServeMux
}

func (m MountContext) WithProtoSecurity(sec pproto.SecurityConfig) MountContext {
	m.ProtoSecurity = sec
	return m
}

func (m MountContext) RegisterNodeCall(callHandler ncall.CallHandler, resolveHandler ncall.ResolveHandler) {
	if m.Host == nil {
		return
	}
	ncall.Register(m.Host, m.NodeSecurity, callHandler, resolveHandler)
}

func (m MountContext) HandleHTTP(path string, fn http.HandlerFunc) {
	if m.AdminMux == nil || fn == nil {
		return
	}
	m.AdminMux.HandleFunc(path, fn)
}

func HandleProto[TReq any, TResp any](m MountContext, protoID coreprotocol.ID, fn func(context.Context, TReq) (TResp, error)) {
	if m.Host == nil {
		return
	}
	pproto.HandleProto[TReq, TResp](m.Host, protoID, m.ProtoSecurity, fn)
}

type NodeCallSegment func(context.Context, ncall.CallContext, ncall.CallReq) (bool, ncall.CallResp, error)

// ChainNodeCall 把多个模块级 node.call 处理片段按顺序串起来。
// 约束：
// - capabilities_show 仍然是基线节点合同，因此这里内建优先处理；
// - 其余模块按显式顺序尝试，谁先声明 handled 就由谁负责；
// - 最终没有模块接管时，统一回 ROUTE_NOT_FOUND。
func ChainNodeCall(showBody func() ncall.CapabilitiesShowBody, segments ...NodeCallSegment) ncall.CallHandler {
	return func(ctx context.Context, meta ncall.CallContext, req ncall.CallReq) (ncall.CallResp, error) {
		if strings.TrimSpace(req.Route) == ncall.RouteNodeV1CapabilitiesShow {
			if showBody == nil {
				return routeNotFoundResp(), nil
			}
			body := showBody()
			return ncall.MarshalProto(&body)
		}
		for _, segment := range segments {
			if segment == nil {
				continue
			}
			handled, resp, err := segment(ctx, meta, req)
			if handled {
				return resp, err
			}
		}
		return routeNotFoundResp(), nil
	}
}

func routeNotFoundResp() ncall.CallResp {
	return ncall.CallResp{Ok: false, Code: string(bfterrors.CodeRouteNotFound), Message: "route not found"}
}
