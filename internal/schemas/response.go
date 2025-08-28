package schemas

type AccessLevel string

const (
	AccessLevelUser  AccessLevel = "user"
	AccessLevelAdmin AccessLevel = "admin"
)

type RegisterResponse struct {
	Name string
	Role AccessLevel
}

type LoginResponse struct {
	Name string //NOTE: do i need it? any of auth related stuff?
}
