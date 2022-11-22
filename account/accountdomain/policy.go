package accountdomain

type PolicyID string

type Policy struct {
	ID       PolicyID
	Name     string
	Policies map[string]int
}
