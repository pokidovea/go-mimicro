package response

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
}

func (response *Response) UnmarshalJSON(data []byte) error {
	var f interface{}
	err := json.Unmarshal(data, &f)
	if err != nil {
		return err
	}

	m := f.(map[string]interface{})

	body, err := processBody(m["body"].(string))
	if err != nil {
		return err
	}
	response.Body = body

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

func processBody(body string) (string, error) {
	matched, err := regexp.MatchString(FILE_PATH_REGEXP, body)
	if err != nil {
		return "", err
	}

	if matched {
		filePath := strings.Replace(body, "file://", "", -1)
		if filePath[0] != '/' {
			configFolder := path.Dir(settings.CONFIG_PATH)
			filePath = path.Join(configFolder, filePath)
		}

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		return string(content), nil
	} else {
		return body, nil
	}

}
