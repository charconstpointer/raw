package raw

import "testing"

func TestGetId(t *testing.T) {
	addr := "127.0.0.1:43256"
	expected := uint32(43256)

	got := getID(addr)
	if got != expected {
		t.Errorf("expecetd %d got %d", expected, got)
	}

}
