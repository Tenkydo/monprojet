package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Structure pour les informations CPU détaillées
type CPUInfo struct {
	VendorID   string `json:"vendor"`
	Family     string `json:"family"`
	Model      string `json:"model"`
	MHz        string `json:"mhz"`
	CacheSize  string `json:"cache_size"`
	UserAgent  string `json:"user_agent"`
}

// Structure pour les données de performance par cœur
type CPUClientCoreData struct {
	Core       int     `json:"core"`
	UserAgent  string  `json:"user_agent"`
	CPUPercent float64 `json:"cpu_percent"`
	Timestamp  string  `json:"timestamp"`
}

// Structure complète reçue du client
type SystemData struct {
	CPUInfo     CPUInfo             `json:"cpu_info"`
	CoreData    []CPUClientCoreData `json:"core_data"`
	Hostname    string              `json:"hostname"`
	OS          string              `json:"os"`
	Platform    string              `json:"platform"`
	CollectedAt string              `json:"collected_at"`
}

// Structure pour l'interface web
type WebData struct {
	Clients map[string]SystemData `json:"clients"`
	LastUpdate string             `json:"last_update"`
}

var clientsData = make(map[string]SystemData)

func main() {
	// Création du dossier de stockage
	if _, err := os.Stat("infoPc"); os.IsNotExist(err) {
		os.Mkdir("infoPc", os.ModePerm)
	}

	// Route pour la page principale
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static2/index.html")
	})

	// Serveur de fichiers statiques
	fs := http.FileServer(http.Dir("static2"))
	http.Handle("/static2/", http.StripPrefix("/static2/", fs))

	// Endpoint pour recevoir les données CPU des agents
	http.HandleFunc("/cpu", func(w http.ResponseWriter, r *http.Request) {
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

		// Stockage en mémoire pour l'interface web
		clientsData[systemData.Hostname] = systemData

		// Sauvegarde sur disque
		filename := filepath.Join("infoPc", fmt.Sprintf("cpu_%s_%d.json", 
			systemData.Hostname, time.Now().UnixNano()))
		
		file, err := os.Create(filename)
		if err != nil {
			log.Printf("❌ Erreur création fichier: %v", err)
			http.Error(w, `{"error":"Impossible de créer le fichier"}`, http.StatusInternalServerError)
			return
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(systemData); err != nil {
			log.Printf("❌ Erreur écriture fichier: %v", err)
		}

		// Logs détaillés
		fmt.Printf("📊 Données reçues de %s:\n", systemData.Hostname)
		fmt.Printf("   🖥️  CPU: %s %s (%s MHz)\n", 
			systemData.CPUInfo.VendorID, 
			systemData.CPUInfo.Model,
			systemData.CPUInfo.MHz)
		fmt.Printf("   💻 OS: %s (%s)\n", systemData.OS, systemData.Platform)
		fmt.Printf("   🧮 Cœurs: %d\n", len(systemData.CoreData))
		
		// Calcul de l'utilisation moyenne
		var totalCPU float64
		for _, core := range systemData.CoreData {
			totalCPU += core.CPUPercent
		}
		avgCPU := totalCPU / float64(len(systemData.CoreData))
		fmt.Printf("   📈 CPU moyen: %.2f%%\n", avgCPU)
		fmt.Printf("   💾 Sauvegardé: %s\n", filename)
		fmt.Println("   ─────────────────────────")

		// Réponse de succès
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"hostname": systemData.Hostname,
			"cores_received": len(systemData.CoreData),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Endpoint pour l'interface web (données en temps réel)
	http.HandleFunc("/api/clients", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		
		webData := WebData{
			Clients: clientsData,
			LastUpdate: time.Now().Format(time.RFC3339),
		}
		
		json.NewEncoder(w).Encode(webData)
	})

	// Endpoint pour les statistiques
	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		stats := map[string]interface{}{
			"total_clients": len(clientsData),
			"clients": make(map[string]interface{}),
		}
		
		for hostname, data := range clientsData {
			var totalCPU float64
			maxCPU := 0.0
			minCPU := 100.0
			
			for _, core := range data.CoreData {
				totalCPU += core.CPUPercent
				if core.CPUPercent > maxCPU {
					maxCPU = core.CPUPercent
				}
				if core.CPUPercent < minCPU {
					minCPU = core.CPUPercent
				}
			}
			
			avgCPU := totalCPU / float64(len(data.CoreData))
			
			stats["clients"].(map[string]interface{})[hostname] = map[string]interface{}{
				"cores": len(data.CoreData),
				"avg_cpu": avgCPU,
				"max_cpu": maxCPU,
				"min_cpu": minCPU,
				"last_seen": data.CollectedAt,
				"cpu_model": data.CPUInfo.Model,
				"os": data.OS,
			}
		}
		
		json.NewEncoder(w).Encode(stats)
	})

	fmt.Println("🚀 Serveur CPU Monitor démarré")
	fmt.Println("=====================================")
	fmt.Println("📡 Port: 8888")
	fmt.Println("🌐 Interface web: http://localhost:8888")
	fmt.Println("📊 API clients: http://localhost:8888/api/clients")
	fmt.Println("📈 API stats: http://localhost:8888/api/stats")
	fmt.Println("💾 Stockage: ./infoPc/")
	fmt.Println("=====================================")

	log.Fatal(http.ListenAndServe(":8888", nil))
}