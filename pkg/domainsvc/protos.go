package domainsvc

import "github.com/libp2p/go-libp2p/core/protocol"

const (
	ProtoResolveNamePaid protocol.ID = "/bsv-transfer/domain/resolve_name_paid/1.0.0"
	ProtoQueryNamePaid   protocol.ID = "/bsv-transfer/domain/query_name_paid/1.0.0"
	ProtoRegisterLock    protocol.ID = "/bsv-transfer/domain/register_lock_paid/1.0.0"
	ProtoRegisterSubmit  protocol.ID = "/bsv-transfer/domain/register_submit/1.0.0"
	ProtoSetTargetPaid   protocol.ID = "/bsv-transfer/domain/set_target_paid/1.0.0"
)

const (
	RouteDomainV1Pricing        = "domain.v1.pricing"
	RouteDomainV1Resolve        = "domain.v1.resolve"
	RouteDomainV1Query          = "domain.v1.query"
	RouteDomainV1Lock           = "domain.v1.lock"
	RouteDomainV1ListOwned      = "domain.v1.list_owned"
	RouteDomainV1SetTarget      = "domain.v1.set_target"
	RouteDomainV1RegisterSubmit = "domain.v1.register_submit"
)
