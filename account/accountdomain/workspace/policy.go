package workspace

type Policy struct {
	ID       PolicyID
	Name     string
	Policies map[string]int
}
