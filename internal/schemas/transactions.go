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

type CreateTransactionResponse struct {
	Type            TransactionType   `json:"type"`
	ItemUUID        string            `json:"item_uuid"`
	TransactionUUID string            `json:"transaction_uuid"`
	Amount          int               `json:"amount"`
	Status          TransactionStatus `json:"status"`
	CreatedAt       int64             `json:"created_at"`
}
