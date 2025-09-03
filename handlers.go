package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"
)

// Page principale
func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static2/index.html")
}

// Réception des données CPU + processus
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

	clientsData[systemData.Hostname] = systemData
	saveSystemData(systemData)
	logSystemData(systemData)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":            "ok",
		"hostname":          systemData.Hostname,
		"cores_received":    len(systemData.CoreData),
		"processes_received": len(systemData.Processes),
		"timestamp":         time.Now().Format(time.RFC3339),
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

// API stats
func handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	stats := computeStats(clientsData)
	json.NewEncoder(w).Encode(stats)
}

// API processus
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

	processes := make([]ProcessInfo, len(systemData.Processes))
	copy(processes, systemData.Processes)
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].CPUPercent > processes[j].CPUPercent
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"hostname":  hostname,
		"processes": processes,
		"count":     len(processes),
		"timestamp": systemData.CollectedAt,
	})
}
