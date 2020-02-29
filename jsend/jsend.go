package jsend

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// This provides a JSend function for MOTH
// https://github.com/omniti-labs/jsend

const (
	Success = "success"
	Fail    = "fail"
	Error   = "error"
)

func Write(w http.ResponseWriter, status, short string, format string, a ...interface{}) {
	resp := struct{
		Status string `json:"status"`
		Data   struct {
			Short       string `json:"short"`
			Description string `json:"description"`
		} `json:"data"`
	}{}
	resp.Status = status
	resp.Data.Short = short
	resp.Data.Description = fmt.Sprintf(format, a...)

	respBytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // RFC2616 makes it pretty clear that 4xx codes are for the user-agent
	w.Write(respBytes)
}
