package metrics

type Metric struct {
	Timestamp     int64   `json:"timestamp"`
	SwitchName    string  `json:"switch_name"`
	BandwidthMbps float64 `json:"bandwidth_mbps"`
	LatencyMs     float64 `json:"latency_ms"`
	PacketErrors  int     `json:"packet_errors"`
}
