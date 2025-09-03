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

// Structure complète pour l'envoi
type SystemData struct {
	CPUInfo     CPUInfo             `json:"cpu_info"`
	CoreData    []CPUClientCoreData `json:"core_data"`
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

	// Informations système
	hostname, os, platform, err := getSystemInfo()
	if err != nil {
		log.Printf("⚠️  Erreur collecte système: %v", err)
	}

	return &SystemData{
		CPUInfo:     cpuInfo,
		CoreData:    coreData,
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

	// Affiche les données collectées
	fmt.Printf("📊 Données collectées:\n%s\n", jsonData)

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
	filename := fmt.Sprintf("cpu_local_%d.json", time.Now().Unix())
	
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

func main() {
	fmt.Println("🚀 Démarrage de l'Agent CPU Client")
	fmt.Println("=====================================")

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
	fmt.Println("=====================================")

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

	// Mode d'exécution
	var mode string
	fmt.Print("\n🎯 Mode d'exécution:\n1) Test unique\n2) Monitoring continu\n3) Sauvegarde locale seulement\nChoix (1-3): ")
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
		
	default:
		fmt.Println("❌ Choix invalide")
	}

	fmt.Println("\n👋 Agent terminé")
}