package nodesvc

import "github.com/libp2p/go-libp2p/core/protocol"

const (
	ProtoNodeCall    protocol.ID = "/bsv-transfer/node/call/1.0.0"
	ProtoNodeResolve protocol.ID = "/bsv-transfer/node/resolve/1.0.0"
)

const (
	RouteNodeV1CapabilitiesShow = "node.v1.capabilities_show"
)

const (
	PaymentSchemePool2of2V1 = "pool_2of2_v1"
	PaymentSchemeChainTxV1  = "chain_tx_v1"
)
