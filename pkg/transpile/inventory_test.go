package transpile

import "testing"

func TestInventory(t *testing.T) {
	fs := newTestFs()
	inv, err := FsInventory(fs)
	if err != nil {
		t.Error(err)
	}
	if c, ok := inv["cat0"]; !ok {
		t.Error("No cat0")
	} else if len(c) != 9 {
		t.Error("Wrong category length", c)
	}
}
