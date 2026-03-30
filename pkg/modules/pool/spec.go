package pool

import (
	"github.com/bsv8/BFTP/pkg/infra/caps"
	ncall "github.com/bsv8/BFTP/pkg/infra/ncall"
)

const (
	InternalAbilityID  = "bftp.pool@1"
	PublicCapabilityID = "pool"
	Version            = uint32(1)
)

func Spec() caps.ModuleSpec {
	return caps.ModuleSpec{
		InternalAbility: InternalAbilityID,
		PublicCapability: &caps.PublicCapability{
			ID:      PublicCapabilityID,
			Version: Version,
		},
		Routes: []string{
			ncall.RoutePoolV1Info,
			ncall.RoutePoolV1Create,
			ncall.RoutePoolV1BaseTx,
			ncall.RoutePoolV1PayConfirm,
			ncall.RoutePoolV1Close,
			ncall.RoutePoolV1SessionState,
		},
	}
}
