package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Stockage en mémoire
var clientsData = make(map[string]SystemData)

// Sauvegarde sur disque
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
