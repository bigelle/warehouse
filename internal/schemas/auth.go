package schemas

import "encoding/json"

type Role int

const (
	RoleUndefined Role = iota - 1
	RoleUser
	RoleStocker
	RoleAdmin
)

func (r Role) MarshalJSON() ([]byte, error) {
	return []byte(`"` + r.String() + `"`), nil
}

func (r *Role) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	role := RoleFromString(str)
	*r = role
	return nil
}

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
	Username string `json:"username"`
	UUID     string `json:"uuid"`
	Role     Role   `json:"role"`
}

type LoginRequest struct {
	Username string `validate:"required" json:"username"`
	Password string `validate:"required" json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	Expires     int64  `json:"expires"`
}
