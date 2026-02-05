package config

const SUBTYPE_KCPFLEET = "kcp-fleet"

type KCPFleetConfig struct {
	EndpointSlice string
}

func (r *KCPFleetConfig) GetType() string {
	return SUBTYPE_KCPFLEET
}
