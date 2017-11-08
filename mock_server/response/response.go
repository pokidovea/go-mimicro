package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/pokidovea/mimicro/settings"
)

const FILE_PATH_REGEXP = `^file:\/\/[\/\w\.]*$`

type Response struct {
	Body        string `json:"body"`
	ContentType string `json:"content_type"`
	StatusCode  int    `json:"status_code"`
	isFile      bool
}

func (response *Response) UnmarshalJSON(data []byte) error {
	var f interface{}
	err := json.Unmarshal(data, &f)
	if err != nil {
		return err
	}

	m := f.(map[string]interface{})

	err = response.setBody(m["body"].(string))
	if err != nil {
		return err
	}

	if val, ok := m["content_type"]; !ok {
		if !response.isFile {
			response.ContentType = "text/plain"
		}
		// otherwise ctype will be detected automatically
	} else {
		response.ContentType = val.(string)
	}

	val, ok := m["status_code"]
	if !ok || response.isFile {
		response.StatusCode = http.StatusOK
	} else {
		response.StatusCode = int(val.(float64))
	}

	return nil
}

func (response *Response) WriteResponse(w http.ResponseWriter, req *http.Request) {
	if response.ContentType != "" {
		w.Header().Set("Content-Type", response.ContentType)
	}

	if response.isFile {
		http.ServeFile(w, req, response.Body)
	} else {
		w.WriteHeader(response.StatusCode)
		fmt.Fprintf(w, response.Body)
	}
}

func (response *Response) setBody(body string) error {
	matched, err := regexp.MatchString(FILE_PATH_REGEXP, body)
	if err != nil {
		return err
	}

	if matched {
		response.isFile = true
		filePath := strings.Replace(body, "file://", "", -1)

		if filePath[0] != '/' {
			configFolder := path.Dir(settings.CONFIG_PATH)
			filePath = path.Join(configFolder, filePath)
			response.Body = filePath
		}
	} else {
		response.Body = body
	}

	return nil
}
