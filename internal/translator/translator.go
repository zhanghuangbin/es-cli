package translator

import "context"

type Meta struct {
	Status  int
	Message string
}

type Result struct {
	Meta    Meta
	Columns []string
	Rows    [][]any
}

type Translator interface {
	Execute(ctx context.Context, sql string) (*Result, error)
}
