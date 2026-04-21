package arbiter

import (
	"github.com/bsv8/BFTP/pkg/infra/caps"
)

const (
	InternalAbilityID = "bftp.arbiter@1"
	PublicCapabilityID = "arbiter"
	Version            = uint32(1)
)

func Spec() caps.ModuleSpec {
	return caps.ModuleSpec{
		InternalAbility: InternalAbilityID,
		Capabilities: []caps.PublicCapability{
			{ID: PublicCapabilityID, Version: Version, ProtocolID: "/bsv-transfer/arbiter/healthz/1.0.0"},
			{ID: PublicCapabilityID, Version: Version, ProtocolID: "/bsv-transfer/arbiter/case/open/1.0.0"},
			{ID: PublicCapabilityID, Version: Version, ProtocolID: "/bsv-transfer/arbiter/case/sign/1.0.0"},
			{ID: PublicCapabilityID, Version: Version, ProtocolID: "/bsv-transfer/arbiter/case/get/1.0.0"},
		},
	}
}
