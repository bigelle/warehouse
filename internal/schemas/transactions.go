package schemas

type TransactionType string

const (
	TransactionTypeRestock  TransactionType = "restock"
	TransactionTypeWithdraw TransactionType = "withdraw"
)

type TransactionStatus string

const (
	TransactionStatusSucceeded TransactionStatus = "succeeded"
	TransactionStatusFailed    TransactionStatus = "failed"
)

type CreateTransactionRequest struct {
	Type     TransactionType `validate:"required, oneof=restock withdraw" json:"type"`
	ItemUUID string          `validate:"required, uuid" json:"item_uuid"`
	Amount   int             `validate:"required, min=1" json:"amount"`
}

type GetAllTransactionsRequest struct {
	Offset int `json:"offset" form:"offset"`
	Limit  int `json:"limit" form:"limit"`
}

type GetAllTransactionsResponse struct {
	NResult      int           `json:"n_result"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	Type      TransactionType   `json:"type"`
	UUID      string            `json:"uuid"`
	OwnerUUID string            `json:"owner_uuid"`
	ItemUUID  string            `json:"item_uuid"`
	Amount    int               `json:"amount"`
	Status    TransactionStatus `json:"status"`
	CreatedAt int64             `json:"created_at"`
}
