package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ========== STRUCTURES (Models) ==========

// Informations CPU
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

// Structure pour les informations de processus
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

// Données système complètes (MISE À JOUR avec processus)
type SystemData struct {
	CPUInfo     CPUInfo             `json:"cpu_info"`
	CoreData    []CPUClientCoreData `json:"core_data"`
	Processes   []ProcessInfo       `json:"processes"`  // NOUVEAU
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

// Statistiques des processus
type ProcessStats struct {
	TopCPUProcesses []ProcessInfo `json:"top_cpu_processes"`
	TopMemProcesses []ProcessInfo `json:"top_memory_processes"`
	TotalProcesses  int           `json:"total_processes"`
	RunningProcs    int           `json:"running_processes"`
	SleepingProcs   int           `json:"sleeping_processes"`
}

// ========== STOCKAGE ==========

var clientsData = make(map[string]SystemData)

// Sauvegarde sur disque (MISE À JOUR)
func saveSystemData(systemData SystemData) {
	filename := filepath.Join("infoPc", fmt.Sprintf("system_%s_%d.json",
		systemData.Hostname, time.Now().UnixNano()))
	
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("❌ Erreur création fichier: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(systemData); err != nil {
		log.Printf("❌ Erreur écriture fichier: %v", err)
	} else {
		fmt.Printf("💾 Sauvegardé: %s (avec %d processus)\n", filename, len(systemData.Processes))
	}
}

// ========== HANDLERS ==========

// Page principale
func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static2/index.html")
}

// Réception des données CPU (MISE À JOUR avec processus)
func handleCPU(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Méthode non autorisée"}`, http.StatusMethodNotAllowed)
		return
	}

	var systemData SystemData
	if err := json.NewDecoder(r.Body).Decode(&systemData); err != nil {
		log.Printf("❌ Erreur décodage JSON: %v", err)
		http.Error(w, `{"error":"Impossible de décoder le JSON"}`, http.StatusBadRequest)
		return
	}

	// Stockage en mémoire
	clientsData[systemData.Hostname] = systemData
	
	// Sauvegarde disque
	saveSystemData(systemData)

	// Logs
	logSystemData(systemData)

	// Réponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":           "ok",
		"hostname":         systemData.Hostname,
		"cores_received":   len(systemData.CoreData),
		"processes_received": len(systemData.Processes), // NOUVEAU
		"timestamp":        time.Now().Format(time.RFC3339),
	})
}

// API clients
func handleClients(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	webData := WebData{
		Clients:    clientsData,
		LastUpdate: time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(webData)
}

// API stats (MISE À JOUR avec statistiques processus)
func handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	stats := computeStats(clientsData)
	json.NewEncoder(w).Encode(stats)
}

// NOUVEAU: API pour les processus détaillés
func handleProcesses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	hostname := r.URL.Query().Get("hostname")
	if hostname == "" {
		http.Error(w, `{"error":"hostname requis"}`, http.StatusBadRequest)
		return
	}

	systemData, exists := clientsData[hostname]
	if !exists {
		http.Error(w, `{"error":"client non trouvé"}`, http.StatusNotFound)
		return
	}

	// Tri des processus par utilisation CPU
	processes := make([]ProcessInfo, len(systemData.Processes))
	copy(processes, systemData.Processes)
	
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].CPUPercent > processes[j].CPUPercent
	})

	response := map[string]interface{}{
		"hostname":  hostname,
		"processes": processes,
		"count":     len(processes),
		"timestamp": systemData.CollectedAt,
	}

	json.NewEncoder(w).Encode(response)
}

// ========== UTILS ==========

// Logs lisibles (MISE À JOUR avec processus)
func logSystemData(systemData SystemData) {
	fmt.Printf("📊 Données reçues de %s:\n", systemData.Hostname)
	fmt.Printf("   🖥️  CPU: %s %s (%s MHz)\n",
		systemData.CPUInfo.VendorID,
		systemData.CPUInfo.Model,
		systemData.CPUInfo.MHz)
	fmt.Printf("   💻 OS: %s (%s)\n", systemData.OS, systemData.Platform)
	fmt.Printf("   🧮 Cœurs: %d\n", len(systemData.CoreData))
	fmt.Printf("   ⚙️  Processus: %d\n", len(systemData.Processes)) // NOUVEAU

	// Calcul CPU moyen
	totalCPU := 0.0
	for _, core := range systemData.CoreData {
		totalCPU += core.CPUPercent
	}
	avgCPU := totalCPU / float64(len(systemData.CoreData))
	fmt.Printf("   📈 CPU moyen: %.2f%%\n", avgCPU)

	// Top 3 processus par CPU
	if len(systemData.Processes) > 0 {
		processes := make([]ProcessInfo, len(systemData.Processes))
		copy(processes, systemData.Processes)
		
		sort.Slice(processes, func(i, j int) bool {
			return processes[i].CPUPercent > processes[j].CPUPercent
		})

		fmt.Printf("   🏆 Top processus:\n")
		limit := 3
		if len(processes) < limit {
			limit = len(processes)
		}
		
		for i := 0; i < limit; i++ {
			proc := processes[i]
			if proc.CPUPercent > 0 {
				fmt.Printf("      %d. %s (PID:%d) - CPU:%.1f%% Mem:%.1f%%\n",
					i+1, proc.Name, proc.PID, proc.CPUPercent, proc.MemPercent)
			}
		}
	}

	fmt.Println("   ─────────────────────────")
}

// Calcul des statistiques globales (MISE À JOUR avec processus)
func computeStats(data map[string]SystemData) map[string]interface{} {
	stats := map[string]interface{}{
		"total_clients": len(data),
		"clients":       make(map[string]interface{}),
		"global_stats":  make(map[string]interface{}),
	}

	totalProcesses := 0
	allProcesses := []ProcessInfo{}

	for hostname, d := range data {
		// Stats CPU existantes
		var totalCPU float64
		maxCPU := 0.0
		minCPU := 100.0
		
		for _, core := range d.CoreData {
			totalCPU += core.CPUPercent
			if core.CPUPercent > maxCPU {
				maxCPU = core.CPUPercent
			}
			if core.CPUPercent < minCPU {
				minCPU = core.CPUPercent
			}
		}
		avgCPU := totalCPU / float64(len(d.CoreData))

		// Stats processus
		processStats := computeProcessStats(d.Processes)
		totalProcesses += len(d.Processes)
		
		// Ajouter tous les processus pour les stats globales
		for _, proc := range d.Processes {
			proc.PID = proc.PID // Garder le PID original mais ajouter le hostname
			allProcesses = append(allProcesses, proc)
		}

		stats["clients"].(map[string]interface{})[hostname] = map[string]interface{}{
			"cores":           len(d.CoreData),
			"avg_cpu":         avgCPU,
			"max_cpu":         maxCPU,
			"min_cpu":         minCPU,
			"last_seen":       d.CollectedAt,
			"cpu_model":       d.CPUInfo.Model,
			"os":              d.OS,
			"processes_count": len(d.Processes), // NOUVEAU
			"process_stats":   processStats,     // NOUVEAU
		}
	}

	// Stats globales des processus
	globalProcessStats := computeProcessStats(allProcesses)
	stats["global_stats"] = map[string]interface{}{
		"total_processes": totalProcesses,
		"process_stats":   globalProcessStats,
	}

	return stats
}

// NOUVEAU: Calcul des statistiques des processus
func computeProcessStats(processes []ProcessInfo) ProcessStats {
	if len(processes) == 0 {
		return ProcessStats{}
	}

	// Tri par CPU
	cpuSorted := make([]ProcessInfo, len(processes))
	copy(cpuSorted, processes)
	sort.Slice(cpuSorted, func(i, j int) bool {
		return cpuSorted[i].CPUPercent > cpuSorted[j].CPUPercent
	})

	// Tri par mémoire
	memSorted := make([]ProcessInfo, len(processes))
	copy(memSorted, processes)
	sort.Slice(memSorted, func(i, j int) bool {
		return memSorted[i].MemPercent > memSorted[j].MemPercent
	})

	// Top 5 par CPU et mémoire
	topCPU := cpuSorted
	if len(topCPU) > 5 {
		topCPU = cpuSorted[:5]
	}

	topMem := memSorted
	if len(topMem) > 5 {
		topMem = memSorted[:5]
	}

	// Comptage des status
	runningCount := 0
	sleepingCount := 0
	for _, proc := range processes {
		switch proc.Status {
		case "R", "running":
			runningCount++
		case "S", "sleeping":
			sleepingCount++
		}
	}

	return ProcessStats{
		TopCPUProcesses: topCPU,
		TopMemProcesses: topMem,
		TotalProcesses:  len(processes),
		RunningProcs:    runningCount,
		SleepingProcs:   sleepingCount,
	}
}

// ========== MAIN ==========

func main() {
	// Création du dossier de stockage
	if _, err := os.Stat("infoPc"); os.IsNotExist(err) {
		_ = os.Mkdir("infoPc", os.ModePerm)
	}

	// Routes
	http.HandleFunc("/", serveIndex)
	http.Handle("/static2/", http.StripPrefix("/static2/", http.FileServer(http.Dir("static2"))))
	http.HandleFunc("/cpu", handleCPU)
	http.HandleFunc("/api/clients", handleClients)
	http.HandleFunc("/api/stats", handleStats)
	http.HandleFunc("/api/processes", handleProcesses) // NOUVEAU

	// Logs serveur
	fmt.Println("🚀 Serveur CPU Monitor avec Support Processus démarré")
	fmt.Println("📡 Port: 8888")
	fmt.Println("🌐 Interface web: http://localhost:8888")
	fmt.Println("📊 API clients: http://localhost:8888/api/clients")
	fmt.Println("📈 API stats: http://localhost:8888/api/stats")
	fmt.Println("⚙️  API processus: http://localhost:8888/api/processes?hostname=CLIENT")
	fmt.Println("💾 Stockage: ./infoPc/")
	fmt.Println("===============================================================")

	log.Fatal(http.ListenAndServe(":8888", nil))
}