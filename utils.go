package main

import (
	"fmt"
)

// Logs lisibles
func logSystemData(systemData SystemData) {
	fmt.Printf("📊 Données reçues de %s:\n", systemData.Hostname)
	fmt.Printf("   🖥️  CPU: %s %s (%s MHz)\n",
		systemData.CPUInfo.VendorID,
		systemData.CPUInfo.Model,
		systemData.CPUInfo.MHz)
	fmt.Printf("   💻 OS: %s (%s)\n", systemData.OS, systemData.Platform)
	fmt.Printf("   🧮 Cœurs: %d\n", len(systemData.CoreData))

	totalCPU := 0.0
	for _, core := range systemData.CoreData {
		totalCPU += core.CPUPercent
	}
	avgCPU := totalCPU / float64(len(systemData.CoreData))
	fmt.Printf("   📈 CPU moyen: %.2f%%\n", avgCPU)
	fmt.Println("   ─────────────────────────")
}

// Calcul des statistiques globales
func computeStats(data map[string]SystemData) map[string]interface{} {
	stats := map[string]interface{}{
		"total_clients": len(data),
		"clients":       make(map[string]interface{}),
	}

	for hostname, d := range data {
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

		stats["clients"].(map[string]interface{})[hostname] = map[string]interface{}{
			"cores":     len(d.CoreData),
			"avg_cpu":   avgCPU,
			"max_cpu":   maxCPU,
			"min_cpu":   minCPU,
			"last_seen": d.CollectedAt,
			"cpu_model": d.CPUInfo.Model,
			"os":        d.OS,
		}
	}
	return stats
}
