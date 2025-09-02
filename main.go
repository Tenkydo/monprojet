package main

import (
	"fmt"
	"monprojet/user"
)

func main() {
	// Création des utilisateurs avec le constructeur
	u1 := user.New("Bob Sinclair", "bob")
	u2 := user.New("Alice Johnson", "alice")

	// Afficher les utilisateurs
	fmt.Println("Utilisateur 1 :", u1.GetName())
	fmt.Println("Utilisateur 2 :", u2.GetName())

	// Mettre à jour le nom du premier utilisateur
	u1.UpdateName("Robert")
	fmt.Println("Utilisateur 1 après modification :", u1.GetName())
}
