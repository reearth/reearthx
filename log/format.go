package log

import "fmt"

type Format struct {
	Format string
	Args   []any
}

func (f Format) IsEmpty() bool {
	return f.Format == "" && len(f.Args) == 0
}

func (f Format) Append(g Format) Format {
	if g.IsEmpty() {
		return f
	}
	return Format{
		Format: f.Format + g.Format,
		Args:   append(f.Args, g.Args...),
	}
}

func (f Format) Prepend(g Format) Format {
	if g.IsEmpty() {
		return f
	}
	return Format{
		Format: g.Format + f.Format,
		Args:   append(g.Args, f.Args...),
	}
}

func (f Format) String() string {
	if len(f.Args) == 0 {
		return f.Format
	}
	if f.Format == "" {
		return fmt.Sprint(f.Args...)
	}
	return fmt.Sprintf(f.Format, f.Args...)
}
