package nodesvc

import (
	"context"

	"github.com/bsv8/BFTP/pkg/infra/pproto"
	"github.com/libp2p/go-libp2p/core/host"
)

type CallContext struct {
	SenderPubkeyHex string
	MessageID       string
}

type CallHandler func(context.Context, CallContext, CallReq) (CallResp, error)
type ResolveHandler func(context.Context, ResolveReq) (ResolveResp, error)

// Register 在任意节点上挂载统一的 node.call / node.resolve 协议。
// 设计说明：
// - 节点之间统一通过 bitfs-node 安全域交换“路由调用”与“路由索引解析”；
// - 业务语义由 route 决定，node 层只负责把 sender/message_id 送给业务处理器；
// - 这样 domain/gateway/client 都能复用同一套 peer.call 外壳，不再各自发明一套协议入口。
func Register(h host.Host, sec p2prpc.SecurityConfig, callHandler CallHandler, resolveHandler ResolveHandler) {
	if h == nil {
		return
	}
	p2prpc.HandleProto[CallReq, CallResp](h, ProtoNodeCall, sec, func(ctx context.Context, req CallReq) (CallResp, error) {
		if callHandler == nil {
			return CallResp{Ok: false, Code: "ROUTE_NOT_FOUND", Message: "route not found"}, nil
		}
		senderPubkeyHex, ok := p2prpc.SenderPubkeyHexFromContext(ctx)
		if !ok {
			return CallResp{Ok: false, Code: "UNAUTHORIZED", Message: "sender identity missing"}, nil
		}
		messageID, ok := p2prpc.MessageIDFromContext(ctx)
		if !ok {
			return CallResp{Ok: false, Code: "BAD_REQUEST", Message: "message id missing"}, nil
		}
		return callHandler(ctx, CallContext{
			SenderPubkeyHex: senderPubkeyHex,
			MessageID:       messageID,
		}, req)
	})
	p2prpc.HandleProto[ResolveReq, ResolveResp](h, ProtoNodeResolve, sec, func(ctx context.Context, req ResolveReq) (ResolveResp, error) {
		if resolveHandler == nil {
			return ResolveResp{Ok: false, Code: "NOT_FOUND", Message: "route not found"}, nil
		}
		return resolveHandler(ctx, req)
	})
}
