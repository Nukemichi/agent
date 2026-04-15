package domain

import "encoding/json"

// DeployRequest - payload for POST /api/v1/deploy
type DeployRequest struct {
	BinaryURL string          `json:"binary_url"`
	Config    json.RawMessage `json:"config"`
}

// StatsResponse - response for GET /api/v1/stats
type StatsResponse struct {
	CPU           float64 `json:"cpu_percent"`
	RAMUsed       uint64  `json:"ram_used_mb"`
	RAMTotal      uint64  `json:"ram_total_mb"`
	DiskUsed      uint64  `json:"disk_used_gb"`
	DiskTotal     uint64  `json:"disk_total_gb"`
	Uptime        uint64  `json:"uptime_seconds"`
	ServiceStatus string  `json:"service_status"`
}
