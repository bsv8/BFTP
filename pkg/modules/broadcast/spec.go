package broadcast

import "github.com/bsv8/BFTP/pkg/infra/caps"

const (
	InternalAbilityID  = "bftp.broadcast@1"
	PublicCapabilityID = "broadcast"
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
			RouteBroadcastV1DemandPublish,
			RouteBroadcastV1DemandPublishBatch,
			RouteBroadcastV1LiveDemandPublish,
			RouteBroadcastV1NodeReachabilityAnnounce,
			RouteBroadcastV1NodeReachabilityQuery,
		},
	}
}
