package mock

import (
	cast "github.com/barnybug/go-cast"
	"golang.org/x/net/context"
)

type Scanner struct {
	ScanFuncCalled int
	ScanFunc       func(ctx context.Context, results chan<- *cast.Client) error
}

func (s *Scanner) Scan(ctx context.Context, results chan<- *cast.Client) error {
	s.ScanFuncCalled++
	return s.ScanFunc(ctx, results)
}
