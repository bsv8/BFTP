package domainmodule

import (
	contractprotoid "github.com/bsv8/BFTP-contract/pkg/v1/protoid"
	contractroute "github.com/bsv8/BFTP-contract/pkg/v1/route"
)

const (
	ProtoResolveNamePaid = contractprotoid.ProtoDomainResolveNamePaid
	ProtoQueryNamePaid   = contractprotoid.ProtoDomainQueryNamePaid
	ProtoRegisterLock    = contractprotoid.ProtoDomainRegisterLock
	ProtoRegisterSubmit  = contractprotoid.ProtoDomainRegisterSubmit
	ProtoSetTargetPaid   = contractprotoid.ProtoDomainSetTargetPaid
)

const (
	RouteDomainV1Pricing        = string(contractroute.RouteDomainV1Pricing)
	RouteDomainV1Resolve        = string(contractroute.RouteDomainV1Resolve)
	RouteDomainV1Query          = string(contractroute.RouteDomainV1Query)
	RouteDomainV1Lock           = string(contractroute.RouteDomainV1Lock)
	RouteDomainV1ListOwned      = string(contractroute.RouteDomainV1ListOwned)
	RouteDomainV1SetTarget      = string(contractroute.RouteDomainV1SetTarget)
	RouteDomainV1RegisterSubmit = string(contractroute.RouteDomainV1RegisterSubmit)
)
