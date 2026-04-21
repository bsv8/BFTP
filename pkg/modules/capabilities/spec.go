package capabilities

import (
	"github.com/bsv8/BFTP/pkg/infra/caps"
	ncall "github.com/bsv8/BFTP/pkg/infra/ncall"
)

const (
	InternalAbilityID = "bftp.capabilities@1"
	Version           = uint32(1)
)

func Spec() caps.ModuleSpec {
	return caps.ModuleSpec{
		InternalAbility: InternalAbilityID,
		Protos: []string{
			string(ncall.ProtoCapabilitiesShow),
		},
	}
}
