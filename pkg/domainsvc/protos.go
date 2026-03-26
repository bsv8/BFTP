package domainsvc

import "github.com/libp2p/go-libp2p/core/protocol"

const (
	ProtoResolveNamePaid protocol.ID = "/bsv-transfer/domain/resolve_name_paid/1.0.0"
	ProtoQueryNamePaid   protocol.ID = "/bsv-transfer/domain/query_name_paid/1.0.0"
	ProtoRegisterLock    protocol.ID = "/bsv-transfer/domain/register_lock_paid/1.0.0"
	ProtoRegisterSubmit  protocol.ID = "/bsv-transfer/domain/register_submit/1.0.0"
	ProtoSetTargetPaid   protocol.ID = "/bsv-transfer/domain/set_target_paid/1.0.0"
)
