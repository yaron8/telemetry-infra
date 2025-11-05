package telemetrics

type MetricRecord struct {
	Timestamp     int64   `json:"timestamp"`
	SwitchID      string  `json:"switch_id"`
	BandwidthMbps float64 `json:"bandwidth_mbps"`
	LatencyMs     float64 `json:"latency_ms"`
	PacketErrors  int     `json:"packet_errors"`
}

func GetCSVHeader() []string {
	return []string{"timestamp", "switch_id", "bandwidth_mbps", "latency_ms", "packet_errors"}
}
