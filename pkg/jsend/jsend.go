package jsend

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// This provides a JSend function for MOTH
// https://github.com/omniti-labs/jsend

const (
	// Success is the return code indicating "All went well, and (usually) some data was returned".
	Success = "success"

	// Fail is the return code indicating "There was a problem with the data submitted, or some pre-condition of the API call wasn't satisfied".
	Fail = "fail"

	// Error is the return code indicating "An error occurred in processing the request, i.e. an exception was thrown".
	Error = "error"
)

// JSONWrite writes out data as JSON, sending headers and content length
func JSONWrite(w http.ResponseWriter, data interface{}) {
	respBytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(respBytes)))
	w.WriteHeader(http.StatusOK) // RFC2616 makes it pretty clear that 4xx codes are for the user-agent
	w.Write(respBytes)
}

// Send sends arbitrary data as a JSend response
func Send(w http.ResponseWriter, status string, data interface{}) {
	resp := struct {
		Status string      `json:"status"`
		Data   interface{} `json:"data"`
	}{}
	resp.Status = status
	resp.Data = data

	JSONWrite(w, resp)
}

// Sendf sends a Sprintf()-formatted string as a JSend response
func Sendf(w http.ResponseWriter, status, short string, format string, a ...interface{}) {
	data := struct {
		Short       string `json:"short"`
		Description string `json:"description"`
	}{}
	data.Short = short
	data.Description = fmt.Sprintf(format, a...)

	Send(w, status, data)
}
