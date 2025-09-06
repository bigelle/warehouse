package schemas

const (
	GetItemsRequestDefaultLimit = 50
)

type GetItemsRequest struct {
	Limit  int `validate:"min=0 max=100" json:"limit"`
	Offset int `validate:"min=0" json:"offset"`
	//TODO: sorting?
}

type Item struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type GetItemsResponse struct {
	NResults int    `json:"n_results"`
	Items    []Item `json:"items"`
}
