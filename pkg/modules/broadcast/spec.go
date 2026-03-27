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
		Protos: []string{
			string(ProtoDemandPublishPaid),
			string(ProtoDemandPublishBatchPaid),
			string(ProtoLiveDemandPublishPaid),
			string(ProtoNodeReachabilityAnnouncePaid),
			string(ProtoNodeReachabilityQueryPaid),
		},
	}
}
