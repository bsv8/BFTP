package ncall

import (
	contractprotoid "github.com/bsv8/BFTP-contract/pkg/v1/protoid"
	contractroute "github.com/bsv8/BFTP-contract/pkg/v1/route"
)

const (
	ProtoNodeCall    = contractprotoid.ProtoNodeCall
	ProtoNodeResolve = contractprotoid.ProtoNodeResolve
)

const (
	RouteNodeV1CapabilitiesShow = string(contractroute.RouteNodeV1CapabilitiesShow)
	RoutePoolV1Info             = string(contractroute.RoutePoolV1Info)
	RoutePoolV1Create           = string(contractroute.RoutePoolV1Create)
	RoutePoolV1BaseTx           = string(contractroute.RoutePoolV1BaseTx)
	RoutePoolV1PayConfirm       = string(contractroute.RoutePoolV1PayConfirm)
	RoutePoolV1Close            = string(contractroute.RoutePoolV1Close)
	RoutePoolV1SessionState     = string(contractroute.RoutePoolV1SessionState)
)

const (
	PaymentSchemePool2of2V1 = contractprotoid.PaymentSchemePool2of2V1
	PaymentSchemeChainTxV1  = contractprotoid.PaymentSchemeChainTxV1
)
