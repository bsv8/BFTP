package caps

import (
	"strings"

	ncall "github.com/bsv8/BFTP/pkg/infra/ncall"
)

const (
	InternalAbilityCapabilitiesV1 = "bftp.capabilities@1"
	InternalAbilityPoolV1         = "bftp.pool@1"
	InternalAbilityBroadcastV1    = "bftp.broadcast@1"
	InternalAbilityDomainV1       = "bftp.domain@1"
	InternalAbilityArbiterV1      = "bftp.arbiter@1"
)

type PublicCapability struct {
	ID      string
	Version uint32
}

type Bundle struct {
	InternalAbilities  []string
	PublicCapabilities []PublicCapability
}

func BuildShowBody(nodePubkeyHex string, items []PublicCapability) ncall.CapabilitiesShowBody {
	body := ncall.CapabilitiesShowBody{
		NodePubkeyHex: strings.ToLower(strings.TrimSpace(nodePubkeyHex)),
		Capabilities:  make([]*ncall.CapabilityItem, 0, len(items)),
	}
	for _, item := range items {
		if strings.TrimSpace(item.ID) == "" {
			continue
		}
		body.Capabilities = append(body.Capabilities, &ncall.CapabilityItem{
			ID:      strings.TrimSpace(item.ID),
			Version: item.Version,
		})
	}
	return body
}
