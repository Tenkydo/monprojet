package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	if _, err := os.Stat("infoPc"); os.IsNotExist(err) {
		_ = os.Mkdir("infoPc", os.ModePerm)
	}

	http.HandleFunc("/", serveIndex)
	http.Handle("/static2/", http.StripPrefix("/static2/", http.FileServer(http.Dir("static2"))))
	http.HandleFunc("/cpu", handleCPU)
	http.HandleFunc("/api/clients", handleClients)
	http.HandleFunc("/api/stats", handleStats)
	http.HandleFunc("/api/processes", handleProcesses)

	fmt.Println("🚀 Serveur CPU Monitor démarré sur :8888")
	log.Fatal(http.ListenAndServe(":8888", nil))
}
