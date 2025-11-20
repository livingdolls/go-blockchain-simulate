package transaction

type Transaction struct {
	From      string
	To        string
	Amount    int64
	Message   string
	Signature string
}
