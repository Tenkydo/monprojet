package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/process"
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

// Structure complète pour l'envoi
type SystemData struct {
	CPUInfo     CPUInfo             `json:"cpu_info"`
	CoreData    []CPUClientCoreData `json:"core_data"`
	Processes   []ProcessInfo       `json:"processes"`
	Hostname    string              `json:"hostname"`
	OS          string              `json:"os"`
	Platform    string              `json:"platform"`
	CollectedAt string              `json:"collected_at"`
}

func getCPUInfo() (CPUInfo, error) {
	cpuInfos, err := cpu.Info()
	if err != nil {
		return CPUInfo{}, err
	}

	if len(cpuInfos) == 0 {
		return CPUInfo{}, fmt.Errorf("aucune information CPU trouvée")
	}

	info := cpuInfos[0]
	
	// Construction de l'User-Agent personnalisé
	hostname, _ := os.Hostname()
	userAgent := fmt.Sprintf("CPUAgent/1.0 (%s; %s %s; %s)",
		hostname,
		runtime.GOOS,
		runtime.GOARCH,
		info.ModelName,
	)

	return CPUInfo{
		VendorID:  info.VendorID,
		Family:    info.Family,
		Model:     info.ModelName,
		MHz:       fmt.Sprintf("%.0f", info.Mhz),
		CacheSize: fmt.Sprintf("%d", info.CacheSize),
		UserAgent: userAgent,
	}, nil
}

func getCPUUsagePerCore() ([]CPUClientCoreData, error) {
	// Obtenir le pourcentage d'utilisation par cœur
	percentages, err := cpu.Percent(time.Second, true)
	if err != nil {
		return nil, err
	}

	hostname, _ := os.Hostname()
	userAgent := fmt.Sprintf("CPUAgent/1.0 (%s; %s %s)",
		hostname,
		runtime.GOOS,
		runtime.GOARCH,
	)

	var coreData []CPUClientCoreData
	for i, percent := range percentages {
		coreData = append(coreData, CPUClientCoreData{
			Core:       i,
			UserAgent:  userAgent,
			CPUPercent: percent,
			Timestamp:  time.Now().Format(time.RFC3339),
		})
	}

	return coreData, nil
}

func getProcessInfo() ([]ProcessInfo, error) {
	fmt.Println("🔍 Collecte des informations des processus...")
	
	// Obtenir la liste de tous les PIDs
	pids, err := process.Pids()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des PIDs: %v", err)
	}

	var processes []ProcessInfo
	
	for _, pid := range pids {
		// Créer une instance de processus
		proc, err := process.NewProcess(pid)
		if err != nil {
			// Processus peut avoir disparu entre temps, on continue
			continue
		}

		// Récupérer les informations du processus
		procInfo := ProcessInfo{PID: pid}
		
		// Nom du processus
		if name, err := proc.Name(); err == nil {
			procInfo.Name = name
		} else {
			procInfo.Name = "unknown"
		}

		// Pourcentage CPU (sur une seconde)
		if cpuPercent, err := proc.CPUPercent(); err == nil {
			procInfo.CPUPercent = cpuPercent
		}

		// Pourcentage mémoire
		if memPercent, err := proc.MemoryPercent(); err == nil {
			procInfo.MemPercent = memPercent
		}

		// Status du processus
		if status, err := proc.Status(); err == nil {
			if len(status) > 0 {
				procInfo.Status = status[0]
			}
		} else {
			procInfo.Status = "unknown"
		}

		// Nom d'utilisateur
		if username, err := proc.Username(); err == nil {
			procInfo.Username = username
		} else {
			procInfo.Username = "unknown"
		}

		// Temps de création
		if createTime, err := proc.CreateTime(); err == nil {
			procInfo.CreateTime = createTime
		}

		// Ligne de commande
		if cmdline, err := proc.Cmdline(); err == nil {
			procInfo.CmdLine = cmdline
		}

		// Nombre de threads
		if numThreads, err := proc.NumThreads(); err == nil {
			procInfo.NumThreads = numThreads
		}

		processes = append(processes, procInfo)
		
		// Limiter à 50 processus pour éviter des payloads trop volumineux
		if len(processes) >= 50 {
			break
		}
	}

	fmt.Printf("📊 %d processus collectés\n", len(processes))
	return processes, nil
}

func getSystemInfo() (string, string, string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	hostInfo, err := host.Info()
	if err != nil {
		return hostname, runtime.GOOS, "unknown", err
	}

	return hostname, hostInfo.OS, hostInfo.Platform, nil
}

func collectSystemData() (*SystemData, error) {
	fmt.Println("🔄 Collecte des informations système...")

	// Informations CPU
	cpuInfo, err := getCPUInfo()
	if err != nil {
		return nil, fmt.Errorf("erreur collecte CPU info: %v", err)
	}

	// Données des cœurs
	coreData, err := getCPUUsagePerCore()
	if err != nil {
		return nil, fmt.Errorf("erreur collecte usage CPU: %v", err)
	}

	// Informations des processus
	processes, err := getProcessInfo()
	if err != nil {
		log.Printf("⚠️  Erreur collecte processus: %v", err)
		processes = []ProcessInfo{} // Continue avec une liste vide
	}

	// Informations système
	hostname, os, platform, err := getSystemInfo()
	if err != nil {
		log.Printf("⚠️  Erreur collecte système: %v", err)
	}

	return &SystemData{
		CPUInfo:     cpuInfo,
		CoreData:    coreData,
		Processes:   processes,
		Hostname:    hostname,
		OS:          os,
		Platform:    platform,
		CollectedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func sendDataToServer(data *SystemData, serverURL string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("erreur sérialisation JSON: %v", err)
	}

	// Affiche un résumé des données collectées
	fmt.Printf("📊 Données collectées:\n")
	fmt.Printf("   - CPU: %s %s\n", data.CPUInfo.VendorID, data.CPUInfo.Model)
	fmt.Printf("   - Cœurs: %d\n", len(data.CoreData))
	fmt.Printf("   - Processus: %d\n", len(data.Processes))
	fmt.Printf("   - Taille JSON: %d bytes\n", len(jsonData))

	// Envoi au serveur
	resp, err := http.Post(serverURL+"/cpu", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("erreur envoi au serveur: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("serveur a répondu avec le code: %d", resp.StatusCode)
	}

	fmt.Printf("✅ Données envoyées avec succès au serveur!\n")
	return nil
}

func saveLocalCopy(data *SystemData) error {
	// Sauvegarde locale optionnelle
	filename := fmt.Sprintf("system_data_%d.json", time.Now().Unix())
	
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("💾 Copie locale sauvegardée: %s\n", filename)
	return nil
}

func displayTopProcesses(processes []ProcessInfo, limit int) {
	if len(processes) == 0 {
		return
	}
	
	fmt.Printf("\n🏆 Top %d processus par CPU:\n", limit)
	fmt.Println("PID\tNom\t\tCPU%%\tMem%%\tStatus\tUser")
	fmt.Println("----\t----\t\t----\t----\t------\t----")
	
	count := 0
	for _, proc := range processes {
		if count >= limit {
			break
		}
		if proc.CPUPercent > 0 {
			fmt.Printf("%d\t%-12s\t%.1f\t%.1f\t%s\t%s\n", 
				proc.PID, proc.Name, proc.CPUPercent, proc.MemPercent, 
				proc.Status, proc.Username)
			count++
		}
	}
}

func main() {
	fmt.Println("🚀 Démarrage de l'Agent CPU Client avec Monitoring des Processus")
	fmt.Println("===============================================================")

	// Configuration
	serverURL := "http://192.168.54.203:8888"
	if len(os.Args) > 1 {
		serverURL = os.Args[1]
	}

	// Paramètres par défaut
	interval := 30 * time.Second
	if len(os.Args) > 2 {
		if intervalSec, err := strconv.Atoi(os.Args[2]); err == nil {
			interval = time.Duration(intervalSec) * time.Second
		}
	}

	fmt.Printf("🌐 Serveur cible: %s\n", serverURL)
	fmt.Printf("⏱️  Intervalle: %v\n", interval)
	fmt.Println("===============================================================")

	// Test initial
	fmt.Println("🧪 Test de collecte initial...")
	data, err := collectSystemData()
	if err != nil {
		log.Fatalf("❌ Erreur lors du test initial: %v", err)
	}

	// Affichage des informations de base
	fmt.Printf("🖥️  CPU: %s %s\n", data.CPUInfo.VendorID, data.CPUInfo.Model)
	fmt.Printf("🏠 Hostname: %s\n", data.Hostname)
	fmt.Printf("💻 OS: %s (%s)\n", data.OS, data.Platform)
	fmt.Printf("🧮 Cœurs détectés: %d\n", len(data.CoreData))
	fmt.Printf("⚙️  Processus collectés: %d\n", len(data.Processes))

	// Afficher le top des processus
	displayTopProcesses(data.Processes, 5)

	// Mode d'exécution
	var mode string
	fmt.Print("\n🎯 Mode d'exécution:\n1) Test unique\n2) Monitoring continu\n3) Sauvegarde locale seulement\n4) Affichage détaillé\nChoix (1-4): ")
	fmt.Scanln(&mode)

	switch mode {
	case "1":
		// Test unique
		fmt.Println("\n🔄 Envoi unique...")
		if err := sendDataToServer(data, serverURL); err != nil {
			log.Printf("❌ Erreur: %v", err)
		}
		
	case "2":
		// Monitoring continu
		fmt.Printf("\n🔄 Démarrage du monitoring continu (intervalle: %v)\n", interval)
		fmt.Println("Appuyez sur Ctrl+C pour arrêter...")
		
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				data, err := collectSystemData()
				if err != nil {
					log.Printf("❌ Erreur collecte: %v", err)
					continue
				}

				displayTopProcesses(data.Processes, 3)

				if err := sendDataToServer(data, serverURL); err != nil {
					log.Printf("❌ Erreur envoi: %v", err)
					// Sauvegarde locale en cas d'erreur réseau
					saveLocalCopy(data)
				}
			}
		}
		
	case "3":
		// Sauvegarde locale seulement
		fmt.Println("\n💾 Sauvegarde locale...")
		if err := saveLocalCopy(data); err != nil {
			log.Printf("❌ Erreur sauvegarde: %v", err)
		}
		
	case "4":
		// Affichage détaillé
		fmt.Println("\n📋 Affichage détaillé des données collectées:")
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(jsonData))
		
	default:
		fmt.Println("❌ Choix invalide")
	}

	fmt.Println("\n👋 Agent terminé")
}