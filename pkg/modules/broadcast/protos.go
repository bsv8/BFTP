package broadcast

import "github.com/libp2p/go-libp2p/core/protocol"

const (
	// 需求广播：带支付的发布接口（发布扣费 + 写入 demand）。
	ProtoDemandPublishPaid protocol.ID = "/bsv-transfer/demand/publish_paid/1.0.0"
	// 批量静态需求广播：一次扣费，批量写入多个 demand。
	ProtoDemandPublishBatchPaid protocol.ID = "/bsv-transfer/demand/publish_batch_paid/1.0.0"
	// 直播需求广播：带支付的发布接口（发布扣费 + 写入 live demand）。
	ProtoLiveDemandPublishPaid protocol.ID = "/bsv-transfer/live/demand/publish_paid/1.0.0"
	// 节点地址声明：带支付的发布接口（地址声明写入目录 + 扣费）。
	ProtoNodeReachabilityAnnouncePaid protocol.ID = "/bsv-transfer/node/reachability/announce_paid/1.0.0"
	// 节点地址查询：带支付的查询接口（查询最新有效声明 + 扣费）。
	ProtoNodeReachabilityQueryPaid protocol.ID = "/bsv-transfer/node/reachability/query_paid/1.0.0"
)
