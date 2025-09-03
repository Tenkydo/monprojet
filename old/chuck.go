package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

// structure qui mappe la réponse JSON
type JokeResponse struct {
	Value string `json:"value"`
}

// GetCategories récupère toutes les catégories de blagues
func GetCategories() ([]string, error) {
	resp, err := http.Get("https://api.chucknorris.io/jokes/categories")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var categories []string
	if err := json.Unmarshal(body, &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

// GetJokeByCategory récupère une blague dans la catégorie donnée
func GetJokeByCategory(category string) (string, error) {
	url := fmt.Sprintf("https://api.chucknorris.io/jokes/random?category=%s", category)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var joke JokeResponse
	if err := json.Unmarshal(body, &joke); err != nil {
		return "", err
	}

	return joke.Value, nil
}

// GetRandomJoke récupère une blague aléatoire (sans catégorie)
func GetRandomJoke() (string, error) {
	resp, err := http.Get("https://api.chucknorris.io/jokes/random")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var joke JokeResponse
	if err := json.Unmarshal(body, &joke); err != nil {
		return "", err
	}

	return joke.Value, nil
}

func main() {
	// Récupérer les catégories
	categories, err := GetCategories()
	if err != nil {
		fmt.Println("Erreur récupération catégories:", err)
		return
	}

	fmt.Println("Catégories disponibles :", categories)

	// Exemple : blague aléatoire
	joke, err := GetRandomJoke()
	if err != nil {
		fmt.Println("Erreur récupération blague :", err)
		return
	}
	fmt.Println("\n😂 Blague aléatoire :")
	fmt.Println(joke)

	// Exemple : blague d'une catégorie aléatoire
	rand.Seed(time.Now().UnixNano())
	randomCategory := categories[rand.Intn(len(categories))]
	jokeCat, err := GetJokeByCategory(randomCategory)
	if err != nil {
		fmt.Println("Erreur récupération blague catégorie :", err)
		return
	}
	fmt.Printf("\n😂 Blague dans la catégorie '%s' :\n%s\n", randomCategory, jokeCat)
}
