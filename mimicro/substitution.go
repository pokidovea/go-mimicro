package mimicro

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
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

func (storage *SubstitutionStorage) add(serverName, URL, method string, response *Response) {
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

func (storage *SubstitutionStorage) del(key keyStruct) {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	delete(storage.substitutions, key)
}

func (storage *SubstitutionStorage) AddSubstitutionHandler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	err = ValidateSchema(body, AddSubstitutionSchema)
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

	storage.add(request.ServerName, request.URL, request.Method, request.Response)

	w.WriteHeader(http.StatusOK)
}

func (storage *SubstitutionStorage) DeleteSubstitutionHandler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	err = ValidateSchema(body, DeleteSubstitutionSchema)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var key keyStruct
	err = json.Unmarshal(body, key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	storage.del(key)

	w.WriteHeader(http.StatusOK)
}
