package discover

import (
	cast "github.com/barnybug/go-cast"
	"golang.org/x/net/context"
)

// Service allows to discover chromecast via multiple means
type Service struct {
	Scanner cast.Scanner
}

// First returns the first chromecast that is discovered by the scanner
func (s Service) First(ctx context.Context) (*cast.Client, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // cancel child-ctx when the first client has been found

	result := make(chan *cast.Client, 1)

	go s.Scanner.Scan(ctx, result)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case client := <-result:
		return client, nil
	}
}

// Named returns the first chromecast that is discovered by the scanner with the given name
func (s Service) Named(ctx context.Context, name string) (*cast.Client, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // cancel child-ctx when the right client has been found

	result := make(chan *cast.Client, 1)

	go s.Scanner.Scan(ctx, result)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case client := <-result:
			if client.Name() == name {
				return client, nil
			}
		}
	}
}

// Uniq forward all client deduplicated
func Uniq(in <-chan *cast.Client, out chan<- *cast.Client) {
	uuids := make(map[string]struct{})
	for c := range in {
		if c == nil {
			continue
		}
		if _, ok := uuids[c.Uuid()]; ok {
			continue
		}
		out <- c
		uuids[c.Uuid()] = struct{}{}
	}
	close(out)
}
