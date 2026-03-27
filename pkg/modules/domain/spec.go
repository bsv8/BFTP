package domainmodule

import "github.com/bsv8/BFTP/pkg/infra/caps"

const (
	InternalAbilityID  = "bftp.domain@1"
	PublicCapabilityID = "domain"
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
			RouteDomainV1Pricing,
			RouteDomainV1Resolve,
			RouteDomainV1Query,
			RouteDomainV1Lock,
			RouteDomainV1ListOwned,
			RouteDomainV1SetTarget,
			RouteDomainV1RegisterSubmit,
		},
	}
}
