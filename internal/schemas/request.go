package schemas

type RegisterRequest struct {
	Username string
	Password string
}

type LoginRequest struct {
	Username string
	Password string
}

var (
	GetItemsRequestDefaultLimit int = 50
)

type GetItemsRequest struct {
	Limit  *int // 0-100, defaults to 50
	Offset int  // no defaults
	// NOTE: can do sorting
}
