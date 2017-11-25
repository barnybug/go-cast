package mdns

import (
	"fmt"
	"strings"
	"time"

	cast "github.com/barnybug/go-cast"
	"github.com/hashicorp/mdns"

	"golang.org/x/net/context"
)

// Scanner uses mdns to scan for chromecasts
type Scanner struct {
	// The chromecasts have 'Timeout' time to reply to each probe.
	Timeout time.Duration
}

// Scan repeatedly scans the network  and synchronously sends the chromecast found into the results channel.
// It finishes when the context is done.
func (s Scanner) Scan(ctx context.Context, results chan<- *cast.Client) error {
	defer close(results)

	// generate entries
	entries := make(chan *mdns.ServiceEntry, 10)
	go func() {
		defer close(entries)
		for {
			if ctx.Err() != nil {
				return
			}
			mdns.Query(&mdns.QueryParam{
				Service: "_googlecast._tcp",
				Domain:  "local",
				Timeout: s.Timeout,
				Entries: entries,
			})
		}
	}()

	// decode entries
	for e := range entries {
		c, err := s.Decode(e)
		if err != nil {
			continue
		}
		select {
		case results <- c:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return ctx.Err()
}

// Decode turns an mdns.ServiceEntry into a cast.Client
func (s Scanner) Decode(entry *mdns.ServiceEntry) (*cast.Client, error) {
	if !strings.Contains(entry.Name, "._googlecast") {
		return nil, fmt.Errorf("fdqn '%s does not contain '._googlecast'", entry.Name)
	}

	client := cast.NewClient(entry.AddrV4, entry.Port)
	info := s.ParseTxtRecord(entry.Info)
	client.SetName(info["fn"])
	client.SetInfo(info)
	return client, nil
}

// ParseTxtRecord a Txt recort into a string map
func (Scanner) ParseTxtRecord(txt string) map[string]string {
	m := make(map[string]string)

	s := strings.Split(txt, "|")
	for _, v := range s {
		s := strings.SplitN(v, "=", 2)
		if len(s) == 2 {
			m[s[0]] = s[1]
		}
	}

	return m
}
