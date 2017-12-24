package mimicro

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/xeipuuv/gojsonschema"
)

type keyStruct struct {
	ServerName string `json:"server_name"`
	URL        string `json:"url"`
	Method     string `json:"method"`
}

type substitutionRequest struct {
	keyStruct
	Response *Response
}

// SubstitutionStorage keeps set of substitutions - alternative responses which can temporary overlap
// default behaviour
type SubstitutionStorage struct {
	substitutions map[keyStruct]*Response
	mutex         sync.RWMutex
}

func NewSubstitutionStorage() *SubstitutionStorage {
	storage := new(SubstitutionStorage)
	storage.substitutions = make(map[keyStruct]*Response)
	return storage
}

func (storage *SubstitutionStorage) makeKey(serverName, URL, method string) keyStruct {
	return keyStruct{ServerName: serverName, URL: URL, Method: method}
}

func (storage *SubstitutionStorage) Add(serverName, URL, method string, response *Response) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	key := storage.makeKey(serverName, URL, method)

	storage.substitutions[key] = response
}

func (storage *SubstitutionStorage) Get(serverName, URL, method string) *Response {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()

	key := storage.makeKey(serverName, URL, method)

	return storage.substitutions[key]
}

func (storage *SubstitutionStorage) Del(serverName, URL, method string) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	key := storage.makeKey(serverName, URL, method)

	delete(storage.substitutions, key)
}

func validateSubstitutionSchema(data []byte) error {
	schemaLoader := gojsonschema.NewStringLoader(substitutionSchema)
	documentLoader := gojsonschema.NewStringLoader(string(data))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil

	}
	var errorString string
	for _, desc := range result.Errors() {
		errorString = fmt.Sprintf("%s%s\n", errorString, desc)
	}

	return errors.New(errorString)
}

func (storage *SubstitutionStorage) AddSubstitutionHandler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	err = validateSubstitutionSchema(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var request substitutionRequest
	err = json.Unmarshal(body, request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	storage.Add(request.ServerName, request.URL, request.Method, request.Response)

	w.WriteHeader(http.StatusOK)
}
