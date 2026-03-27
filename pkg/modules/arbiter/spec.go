package arbiter

import "github.com/bsv8/BFTP/pkg/infra/caps"

const (
	InternalAbilityID  = "bftp.arbiter@1"
	PublicCapabilityID = "arbiter"
	Version            = uint32(1)
)

func Spec() caps.ModuleSpec {
	return caps.ModuleSpec{
		InternalAbility: InternalAbilityID,
		PublicCapability: &caps.PublicCapability{
			ID:      PublicCapabilityID,
			Version: Version,
		},
	}
}
