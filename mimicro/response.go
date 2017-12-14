package mimicro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
)

const filePathRegexp = `^file:\/\/[\/\w\.]*$`

// Response struct contains the information about response, such as content, ctype, status code etc.
type Response struct {
	template   *template.Template
	file       *template.Template
	StatusCode int         `json:"status_code"`
	Headers    http.Header `json:"headers"`
}

// UnmarshalJSON used by json lib. Describes how to translate json config into struct
func (response *Response) UnmarshalJSON(data []byte) error {
	var f interface{}
	err := json.Unmarshal(data, &f)
	if err != nil {
		return err
	}

	m := f.(map[string]interface{})

	if val, ok := m["file"]; ok {
		err = response.setFile(val.(string))
		if err != nil {
			return err
		}
	}
	if val, ok := m["template"]; ok {
		err = response.setTemplate(val.(string))
		if err != nil {
			return err
		}
	}

	response.Headers = http.Header{}
	if m["headers"] != nil {
		for header, value := range m["headers"].(map[string]interface{}) {
			switch v := value.(type) {
			case float64:
				response.Headers.Set(header, strconv.FormatFloat(v, 'f', -1, 64))
			case string:
				response.Headers.Set(header, v)
			}
		}
	}

	if val, ok := m["status_code"]; !ok || response.file != nil {
		response.StatusCode = http.StatusOK
	} else {
		response.StatusCode = int(val.(float64))
	}

	return nil
}

// WriteResponse sends the response to the client according to the response params
func (response *Response) WriteResponse(w http.ResponseWriter, req *http.Request) {
	for header, value := range response.Headers {
		w.Header().Set(header, value[0])
	}

	vars := mux.Vars(req)

	if response.template != nil {
		w.WriteHeader(response.StatusCode)

		if err := response.template.Execute(w, vars); err != nil {
			fmt.Fprintf(w, err.Error())
		}
	} else {
		filePath := bytes.NewBufferString("")
		if err := response.file.Execute(filePath, vars); err != nil {
			fmt.Fprintf(w, err.Error())
		}
		http.ServeFile(w, req, filePath.String())
	}
}

func processFilePath(filePath string, checkExistence bool) (string, error) {
	filePath = strings.Replace(filePath, "file://", "", -1)

	if filePath[0] != '/' {
		configFolder := path.Dir(ConfigPath)
		filePath = path.Join(configFolder, filePath)
	}

	if checkExistence {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return "", fmt.Errorf("File does not exist %s", filePath)
		}
	}

	return filePath, nil
}

func (response *Response) setFile(filePath string) error {
	filePath, err := processFilePath(filePath, false)
	if err != nil {
		return err
	}

	templateInstance := template.New("template")
	_, err = templateInstance.Parse(filePath)
	if err != nil {
		return err
	}

	response.file = templateInstance
	return nil
}

func (response *Response) setTemplate(templateString string) error {
	matched, err := regexp.MatchString(filePathRegexp, templateString)
	if err != nil {
		return err
	}

	templateInstance := template.New("template")
	if matched {
		filePath, err := processFilePath(templateString, true)
		if err != nil {
			return err
		}

		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}
		templateString = string(data)
	}

	_, err = templateInstance.Parse(templateString)
	if err != nil {
		return err
	}

	response.template = templateInstance

	return nil
}
