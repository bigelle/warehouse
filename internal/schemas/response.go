package schemas

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

type RegisterResponse struct {
	Name string
	Role Role
}

type LoginResponse struct {
	AccessToken string
}
