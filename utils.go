package main

import (
	"fmt"
	"sort"
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
	fmt.Printf("   ⚙️  Processus: %d\n", len(systemData.Processes))

	totalCPU := 0.0
	for _, core := range systemData.CoreData {
		totalCPU += core.CPUPercent
	}
	avgCPU := totalCPU / float64(len(systemData.CoreData))
	fmt.Printf("   📈 CPU moyen: %.2f%%\n", avgCPU)

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

// Stats globales
func computeStats(data map[string]SystemData) map[string]interface{} {
	// même logique que ton computeStats original
	// ...
	return nil // à compléter ici si nécessaire
}

// Stats processus
func computeProcessStats(processes []ProcessInfo) ProcessStats {
	// même logique que ton computeProcessStats original
	return ProcessStats{}
}
