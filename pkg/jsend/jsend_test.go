package jsend

import (
	"net/http/httptest"
	"testing"
)

func TestEverything(t *testing.T) {
	w := httptest.NewRecorder()

	Sendf(w, Success, "You have cows", "You have %d cows", 12)
	if w.Result().StatusCode != 200 {
		t.Errorf("HTTP Status code: %d", w.Result().StatusCode)
	}
	if w.Body.String() != `{"status":"success","data":{"short":"You have cows","description":"You have 12 cows"}}` {
		t.Errorf("HTTP Body %s", w.Body.Bytes())
	}
}
