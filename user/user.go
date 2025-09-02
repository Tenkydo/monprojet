package user

// User représente un utilisateur
type User struct {
	Name  string
	Login string
}

// New crée un nouvel utilisateur et retourne un pointeur vers User
func New(name, login string) *User {
	return &User{
		Name:  name,
		Login: login,
	}
}

// UpdateName met à jour le nom de l'utilisateur
func (u *User) UpdateName(name string) {
	u.Name = name
}

// GetName retourne le nom et le login sous forme de chaîne
func (u *User) GetName() string {
	return u.Name + " -> " + u.Login
}
