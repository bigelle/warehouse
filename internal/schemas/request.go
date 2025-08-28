package schemas

type RegisterRequest struct {
	//TODO: something something
	Name     string
	Password string
}

type LoginRequest struct {
	Name     string
	Password string
}
