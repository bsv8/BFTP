package ncall

import "github.com/libp2p/go-libp2p/core/protocol"

const (
	ProtoNodeCall    protocol.ID = "/bsv-transfer/node/call/1.0.0"
	ProtoNodeResolve protocol.ID = "/bsv-transfer/node/resolve/1.0.0"
)

const (
	RouteNodeV1CapabilitiesShow = "node.v1.capabilities_show"
	RoutePoolV1Info             = "pool.v1.info"
	RoutePoolV1Create           = "pool.v1.create"
	RoutePoolV1BaseTx           = "pool.v1.base_tx"
	RoutePoolV1PayConfirm       = "pool.v1.pay_confirm"
	RoutePoolV1Close            = "pool.v1.close"
	RoutePoolV1SessionState     = "pool.v1.session_state"
)

const (
	PaymentSchemePool2of2V1 = "pool_2of2_v1"
	PaymentSchemeChainTxV1  = "chain_tx_v1"
)
