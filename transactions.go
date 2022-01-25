package fpgrowth

type Transaction struct {
	ID    int
	Items []string // slice of item names
}

func (t *Transaction) Exists(item string) bool {
	for _, i := range t.Items {
		if i == item {
			return true
		}
	}
	return false
}
