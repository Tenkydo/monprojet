package main

import (
	"fmt"
	"log"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

// List des CPU = Usage
func GetCPUInfo() {
	// Informations CPU
	cpuInfos, err := cpu.Info()
	if err != nil {
		log.Fatalln("get cpu failed ! : ", err.Error())
	}
	
	for _, ci := range cpuInfos {
		fmt.Println(ci)
	}
}

func main() {
	// Récupération des caractéristiques du processeur
	GetCPUInfo()
	
	// Occupation CPU par seconde
	for {
		percent, _ := cpu.Percent(time.Second, false)
		fmt.Println("Cpu Percent : ", percent)
	}
}