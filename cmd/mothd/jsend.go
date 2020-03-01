package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// This provides a JSend function for MOTH
// https://github.com/omniti-labs/jsend

const (
	JSendSuccess = "success"
	JSendFail    = "fail"
	JSendError   = "error"
)

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

func JSend(w http.ResponseWriter, status string, data interface{}) {
	resp := struct{
		Status string `json:"status"`
		Data   interface{} `json:"data"`
	}{}
	resp.Status = status
	resp.Data = data

	JSONWrite(w, resp)
}

func JSendf(w http.ResponseWriter, status, short string, format string, a ...interface{}) {
	data := struct{
		Short       string `json:"short"`
		Description string `json:"description"`
	}{}
	data.Short = short
	data.Description = fmt.Sprintf(format, a...)
	
	JSend(w, status, data)
}
