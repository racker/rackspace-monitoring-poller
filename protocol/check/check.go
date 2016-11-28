package check

import (
	"encoding/json"
)

type CheckHeader struct {
	Id             string            `json:"id"`
	CheckType      string            `json:"type"`
	Period         uint64            `json:"period"`
	Timeout        uint64            `json:"timeout"`
	EntityId       string            `json:"entity_id"`
	ZoneId         string            `json:"zone_id"`
	Disabled       bool              `json:"disabled"`
	IpAddresses    map[string]string `json:"ip_addresses"`
	TargetAlias    *string           `json:"target_alias"`
	TargetHostname *string           `json:"target_hostname"`
	TargetResolver *string           `json:"target_resolver"`
}

// CheckIn is used for unmarshalling received check requests.
// Since the details are polymorphic, they are captured for delayed unmarshalling via the RawDetails field.
type CheckIn struct {
	CheckHeader

	RawDetails *json.RawMessage `json:"details"`
}
