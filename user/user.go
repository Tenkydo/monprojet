package user

import "fmt"

// User représente un utilisateur
type User struct {
	Name  string // public
	Login string // public
	email string // privé
	age   int    // privé
}

// New crée un nouvel utilisateur et retourne un pointeur vers User
func New(name, login, email string, age int) *User {
	return &User{
		Name:  name,
		Login: login,
		email: email,
		age:   age,
	}
}

// UpdateName met à jour le nom de l'utilisateur
func (u *User) UpdateName(name string) {
	u.Name = name
}

// --- GESTION EMAIL ---
func (u *User) GetEmail() string {
	return u.email
}

func (u *User) UpdateEmail(email string) {
	u.email = email
}

// --- GESTION AGE ---
func (u *User) GetAge() int {
	return u.age
}

func (u *User) UpdateAge(age int) {
	u.age = age
}

// Retourne toutes les infos
func (u *User) GetInfo() string {
	return fmt.Sprintf("%s (%s), %s, âge: %d", u.Name, u.Login, u.email, u.age)
}
