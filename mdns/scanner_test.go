package mdns_test

import (
	"fmt"
	"testing"

	cast "github.com/barnybug/go-cast"
	"github.com/barnybug/go-cast/mdns"
)

var _ cast.Scanner = mdns.Scanner{}

func TestDecodeTxtRecord(t *testing.T) {
	s := mdns.Scanner{}

	txt := `id=87cf98a003f1f1dbd2efe6d19055a617|ve=04|md=Chromecast|ic=/setup/icon.png|fn=Chromecast PO|ca=5|st=0|bs=FA8FCA7EE8A9|rs=`
	exp := map[string]string{
		"id": "87cf98a003f1f1dbd2efe6d19055a617",
		"ve": "04",
		"md": "Chromecast",
		"ic": "/setup/icon.png",
		"fn": "Chromecast PO",
		"ca": "5",
		"st": "0",
		"bs": "FA8FCA7EE8A9",
		"rs": "",
	}

	res := s.ParseTxtRecord(txt)
	if !mapEqual(exp, res) {
		t.Errorf("expected %s; found %s", exp, res)
	}
}

func mapEqual(m1, m2 map[string]string) bool {
	if m1 == nil {
		return m2 == nil
	}
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		if v2, ok := m2[k]; !ok || v1 != v2 {
			fmt.Println(k, v1, v2, ok)
			return false
		}
	}
	return true
}
