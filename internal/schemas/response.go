package schemas

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

type Role int

const (
	RoleUndefined Role = -1
	RoleUser      Role = iota
	RoleAdmin
)

func (r Role) String() string {
	switch r {
	case RoleUser:
		return "user"
	case RoleAdmin:
		return "admin"
	default:
		return "undefined"
	}
}

func RoleFromString(str string) Role {
	switch str {
	case "user":
		return RoleUser
	case "admin":
		return RoleAdmin
	default:
		return RoleUndefined
	}
}

var (
	valid *validator.Validate
	once  = &sync.Once{}
)

func Validator() *validator.Validate {
	once.Do(func() {
		valid = validator.New()
	})
	return valid
}

type RegisterResponse struct {
	Name string
	Role Role
}

type LoginResponse struct {
	AccessToken string
}

type GetItemsResponse struct {
	NResults int
	Items    []Item
}

type Item struct {
	ID       string
	Name     string
	Quantity int
}
