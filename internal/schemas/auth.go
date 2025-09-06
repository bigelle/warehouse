package schemas

type Role int

const (
	RoleUndefined Role = iota - 1
	RoleUser
	RoleStocker
	RoleAdmin
)

func (r Role) String() string {
	if r <= RoleUndefined {
		return ""
	}
	return [3]string{"user", "stocker", "admin"}[r]
}

func RoleFromString(str string) Role {
	switch str {
	case "user":
		return RoleUser
	case "stocker":
		return RoleStocker
	case "admin":
		return RoleAdmin
	default:
		return RoleUndefined
	}
}

type RegisterRequest struct {
	Username string `validate:"required" json:"username"`
	Password string `validate:"required" json:"password"`
	Role     Role   `validate:"omitempty,oneof=user stocker admin" json:"role"`
}

type RegisterResponse struct {
	Username string
	UUID     string
	Role     Role
}

type LoginRequest struct {
	Username string `validate:"required" json:"username"`
	Password string `validate:"required" json:"password"`
}

type LoginResponse struct {
	AccessToken string
	Expires     int64
}
