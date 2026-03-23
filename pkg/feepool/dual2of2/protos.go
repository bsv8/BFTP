package dual2of2

import "github.com/libp2p/go-libp2p/core/protocol"

const (
	// 费用池参数查询（握手参数下发）。
	ProtoFeePoolInfo protocol.ID = "/bsv-transfer/fee_pool/info/1.0.0"

	// Open 阶段：Create + BaseTx。
	ProtoFeePoolCreate protocol.ID = "/bsv-transfer/fee_pool/create/1.0.0"
	ProtoFeePoolBaseTx protocol.ID = "/bsv-transfer/fee_pool/base_tx/1.0.0"

	// Pay 阶段：更新 spend tx（不上链）。
	ProtoFeePoolPayConfirm protocol.ID = "/bsv-transfer/fee_pool/pay_confirm/1.0.0"

	// Close 阶段：最终结算并广播 final tx。
	ProtoFeePoolClose protocol.ID = "/bsv-transfer/fee_pool/close/1.0.0"

	// 状态查询（用于观测/e2e）。
	ProtoFeePoolState protocol.ID = "/bsv-transfer/fee_pool/state/1.0.0"

	// 需求广播：带支付的发布接口（发布扣费 + 写入 demand）。
	ProtoDemandPublishPaid protocol.ID = "/bsv-transfer/demand/publish_paid/1.0.0"
	// 批量静态需求广播：一次扣费，批量写入多个 demand。
	ProtoDemandPublishBatchPaid protocol.ID = "/bsv-transfer/demand/publish_batch_paid/1.0.0"
	// 直播需求广播：带支付的发布接口（发布扣费 + 写入 live demand）。
	ProtoLiveDemandPublishPaid protocol.ID = "/bsv-transfer/live/demand/publish_paid/1.0.0"
)
