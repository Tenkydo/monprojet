package main

import (
	"fmt"
	"github.com/Tenkydo/monprojet/user"
)

func main() {
	u1 := user.New("Bob Sinclair", "bob", "bob@example.com", 35)

	fmt.Println("Utilisateur :", u1.GetInfo())

	// On met à jour email et âge via les setters
	u1.UpdateEmail("bob.sinclair@gmail.com")
	u1.UpdateAge(36)

	// On récupère les nouvelles valeurs via les getters
	fmt.Println("Nouvel email :", u1.GetEmail())
	fmt.Println("Nouvel âge   :", u1.GetAge())
}
