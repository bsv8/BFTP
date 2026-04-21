package pool

import (
	contractprotoid "github.com/bsv8/BFTP-contract/pkg/v1/protoid"
	"github.com/bsv8/BFTP/pkg/infra/caps"
)

const (
	InternalAbilityID = "bftp.pool@1"
	PublicCapabilityID = "pool"
	Version            = uint32(1)
)

func Spec() caps.ModuleSpec {
	return caps.ModuleSpec{
		InternalAbility: InternalAbilityID,
		Capabilities: []caps.PublicCapability{
			{ID: PublicCapabilityID, Version: Version, ProtocolID: string(contractprotoid.ProtoPoolV1Info)},
			{ID: PublicCapabilityID, Version: Version, ProtocolID: string(contractprotoid.ProtoPoolV1Create)},
			{ID: PublicCapabilityID, Version: Version, ProtocolID: string(contractprotoid.ProtoPoolV1BaseTx)},
			{ID: PublicCapabilityID, Version: Version, ProtocolID: string(contractprotoid.ProtoPoolV1PayConfirm)},
			{ID: PublicCapabilityID, Version: Version, ProtocolID: string(contractprotoid.ProtoPoolV1Close)},
			{ID: PublicCapabilityID, Version: Version, ProtocolID: string(contractprotoid.ProtoPoolV1SessionState)},
		},
		Protos: []string{
			string(contractprotoid.ProtoPoolV1Info),
			string(contractprotoid.ProtoPoolV1Create),
			string(contractprotoid.ProtoPoolV1BaseTx),
			string(contractprotoid.ProtoPoolV1PayConfirm),
			string(contractprotoid.ProtoPoolV1Close),
			string(contractprotoid.ProtoPoolV1SessionState),
		},
	}
}
