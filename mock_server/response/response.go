package response

import (
	// "errors"
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Body        string `json:"body"`
	ContentType string `json:"content_type"`
	StatusCode  int    `json:"status_code"`
}

func (response *Response) UnmarshalJSON(data []byte) error {
	var f interface{}
	err := json.Unmarshal(data, &f)
	if err != nil {
		return err
	}

	m := f.(map[string]interface{})

	response.Body = m["body"].(string)

	if val, ok := m["content_type"]; !ok {
		response.ContentType = "text/plain"
	} else {
		response.ContentType = val.(string)
	}

	if val, ok := m["status_code"]; !ok {
		response.StatusCode = http.StatusOK
	} else {
		response.StatusCode = int(val.(float64))
	}

	return nil
}

func (response Response) WriteResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", response.ContentType)
	w.WriteHeader(response.StatusCode)

	fmt.Fprintf(w, response.Body)
}
