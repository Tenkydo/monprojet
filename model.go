package main

// Informations CPU
type CPUInfo struct {
	VendorID  string `json:"vendor"`
	Family    string `json:"family"`
	Model     string `json:"model"`
	MHz       string `json:"mhz"`
	CacheSize string `json:"cache_size"`
	UserAgent string `json:"user_agent"`
}

// Données par cœur
type CPUClientCoreData struct {
	Core       int     `json:"core"`
	UserAgent  string  `json:"user_agent"`
	CPUPercent float64 `json:"cpu_percent"`
	Timestamp  string  `json:"timestamp"`
}

// Données système complètes
type SystemData struct {
	CPUInfo     CPUInfo             `json:"cpu_info"`
	CoreData    []CPUClientCoreData `json:"core_data"`
	Hostname    string              `json:"hostname"`
	OS          string              `json:"os"`
	Platform    string              `json:"platform"`
	CollectedAt string              `json:"collected_at"`
}

// Données pour interface web
type WebData struct {
	Clients    map[string]SystemData `json:"clients"`
	LastUpdate string                `json:"last_update"`
}
