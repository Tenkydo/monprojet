package main

// CPU
type CPUInfo struct {
	VendorID   string `json:"vendor"`
	Family     string `json:"family"`
	Model      string `json:"model"`
	MHz        string `json:"mhz"`
	CacheSize  string `json:"cache_size"`
	UserAgent  string `json:"user_agent"`
}

// Données par cœur
type CPUClientCoreData struct {
	Core       int     `json:"core"`
	UserAgent  string  `json:"user_agent"`
	CPUPercent float64 `json:"cpu_percent"`
	Timestamp  string  `json:"timestamp"`
}

// Infos processus
type ProcessInfo struct {
	PID         int32   `json:"pid"`
	Name        string  `json:"name"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemPercent  float32 `json:"memory_percent"`
	Status      string  `json:"status"`
	Username    string  `json:"username"`
	CreateTime  int64   `json:"create_time"`
	CmdLine     string  `json:"cmdline"`
	NumThreads  int32   `json:"num_threads"`
}

// Données complètes d’un client
type SystemData struct {
	CPUInfo     CPUInfo             `json:"cpu_info"`
	CoreData    []CPUClientCoreData `json:"core_data"`
	Processes   []ProcessInfo       `json:"processes"`
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

// Statistiques processus
type ProcessStats struct {
	TopCPUProcesses []ProcessInfo `json:"top_cpu_processes"`
	TopMemProcesses []ProcessInfo `json:"top_memory_processes"`
	TotalProcesses  int           `json:"total_processes"`
	RunningProcs    int           `json:"running_processes"`
	SleepingProcs   int           `json:"sleeping_processes"`
}
