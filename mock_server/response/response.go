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
	w.Header().Set("Content-Type", response.ContentType)
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, response.Body)
}
