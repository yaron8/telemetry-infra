package telemetrics

type MetricRecord struct {
	SwitchID      string  `json:"switch_id,omitempty"`
	BandwidthMbps float64 `json:"bandwidth_mbps"`
	LatencyMs     float64 `json:"latency_ms"`
	PacketErrors  int     `json:"packet_errors"`
}

func GetCSVHeader() []string {
	return []string{
		"switch_id",
		"bandwidth_mbps",
		"latency_ms",
		"packet_errors"}
}
