package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Page principale
func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static2/index.html")
}

// Réception des données CPU
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

	// Logs
	logSystemData(systemData)

	// Réponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "ok",
		"hostname":       systemData.Hostname,
		"cores_received": len(systemData.CoreData),
		"timestamp":      time.Now().Format(time.RFC3339),
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

	stats := computeStats(clientsData)
	json.NewEncoder(w).Encode(stats)
}
