package usecase

type Cursor string

func CursorFromRef(c *string) *Cursor {
	if c == nil {
		return nil
	}
	d := Cursor(*c)
	return &d
}

func (c Cursor) Ref() *Cursor {
	return &c
}

func (c *Cursor) CopyRef() *Cursor {
	if c == nil {
		return nil
	}
	d := *c
	return &d
}

func (c *Cursor) StringRef() *string {
	if c == nil {
		return nil
	}
	s := string(*c)
	return &s
}
