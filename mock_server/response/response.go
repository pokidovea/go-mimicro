package response

import (
	"fmt"
	"net/http"
)

type Response struct {
	Body        string `json:"body"`
	ContentType string `json:"content_type"`
	StatusCode  int    `json:"status_code"`
}

func (response Response) WriteResponse(w http.ResponseWriter) {
	if response.ContentType == "" {
		w.Header().Set("Content-Type", "text/plain")
	} else {
		w.Header().Set("Content-Type", response.ContentType)
	}

	if response.StatusCode == 0 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(response.StatusCode)
	}

	fmt.Fprintf(w, response.Body)
}
