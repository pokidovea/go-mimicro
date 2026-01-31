package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"mimicro/internal/generator"
	"mimicro/internal/spec"
	"mimicro/internal/validator"
)

type Server struct {
	spec      *spec.OpenAPISpec
	port      int
	generator *generator.ResponseGenerator
	validator *validator.Validator
}

func New(apiSpec *spec.OpenAPISpec, port int) *Server {
	return &Server{
		spec:      apiSpec,
		port:      port,
		generator: generator.New(apiSpec),
		validator: validator.New(apiSpec),
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register all paths from OpenAPI spec
	for path, pathItem := range s.spec.Paths {
		s.registerPath(mux, path, pathItem)

		// Register a catch-all handler for 405 Method Not Allowed
		mux.HandleFunc(path, s.handleMethodNotAllowed(path, pathItem))
	}

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Mock server listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) registerPath(mux *http.ServeMux, path string, pathItem spec.PathItem) {
	if pathItem.Get != nil {
		mux.HandleFunc(fmt.Sprintf("GET %s", path), s.handleRequest(path, "GET", pathItem.Get))
		log.Printf("Registered: GET %s", path)
	}
	if pathItem.Post != nil {
		mux.HandleFunc(fmt.Sprintf("POST %s", path), s.handleRequest(path, "POST", pathItem.Post))
		log.Printf("Registered: POST %s", path)
	}
	if pathItem.Put != nil {
		mux.HandleFunc(fmt.Sprintf("PUT %s", path), s.handleRequest(path, "PUT", pathItem.Put))
		log.Printf("Registered: PUT %s", path)
	}
	if pathItem.Delete != nil {
		mux.HandleFunc(fmt.Sprintf("DELETE %s", path), s.handleRequest(path, "DELETE", pathItem.Delete))
		log.Printf("Registered: DELETE %s", path)
	}
	if pathItem.Patch != nil {
		mux.HandleFunc(fmt.Sprintf("PATCH %s", path), s.handleRequest(path, "PATCH", pathItem.Patch))
		log.Printf("Registered: PATCH %s", path)
	}
}

func (s *Server) handleMethodNotAllowed(path string, pathItem spec.PathItem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// This handler is only called if no specific method handler matched
		allowed := s.getAllowedMethods(pathItem)
		if len(allowed) > 0 {
			log.Printf("405 Method Not Allowed: %s %s from %s", r.Method, path, r.RemoteAddr)
			w.Header().Set("Allow", joinMethods(allowed))
			w.WriteHeader(http.StatusMethodNotAllowed)
			err := json.NewEncoder(w).Encode(map[string]string{
				"error": fmt.Sprintf("Method %s not allowed. Allowed methods: %s", r.Method, joinMethods(allowed)),
			})
			if err != nil {
				log.Printf("Error encoding error response: %v", err)
				return
			}
		}
	}
}

func (s *Server) getAllowedMethods(pathItem spec.PathItem) []string {
	var methods []string
	if pathItem.Get != nil {
		methods = append(methods, "GET")
	}
	if pathItem.Post != nil {
		methods = append(methods, "POST")
	}
	if pathItem.Put != nil {
		methods = append(methods, "PUT")
	}
	if pathItem.Delete != nil {
		methods = append(methods, "DELETE")
	}
	if pathItem.Patch != nil {
		methods = append(methods, "PATCH")
	}
	return methods
}

func joinMethods(methods []string) string {
	result := ""
	for i, m := range methods {
		if i > 0 {
			result += ", "
		}
		result += m
	}
	return result
}

func (s *Server) handleRequest(path, method string, operation *spec.Operation) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s from %s", method, path, r.RemoteAddr)

		// Validate path parameters
		if err := s.validator.ValidatePathParameters(r, operation); err != nil {
			log.Printf("Validation error (path params): %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			if err != nil {
				log.Printf("Error encoding error response: %v", err)
				return
			}
			return
		}

		// Validate query parameters
		if err := s.validator.ValidateQueryParameters(r, operation); err != nil {
			log.Printf("Validation error (query params): %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			if err != nil {
				log.Printf("Error encoding error response: %v", err)
				return
			}
			return
		}

		// Validate request body
		// Need to read body and restore it for validation
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			err := r.Body.Close()
			if err != nil {
				log.Printf("Failed to close request body: %v", err)
				return
			}
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if err := s.validator.ValidateRequestBody(r, operation); err != nil {
			log.Printf("Validation error (request body): %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			if err != nil {
				log.Printf("Error encoding error response: %v", err)
				return
			}
			return
		}

		// Find successful response (200, 201, etc.)
		var response *spec.Response
		var statusCode int

		for code, resp := range operation.Responses {
			if code[0] == '2' { // 2xx success codes
				response = &resp
				_, err := fmt.Sscanf(code, "%d", &statusCode)
				if err != nil {
					log.Printf("Failed to parse response status code: %v", err)
					return
				}
				break
			}
		}

		if response == nil {
			// No success response defined, return 200 OK with empty response
			w.WriteHeader(http.StatusOK)
			return
		}

		// Generate response based on content type
		for contentType, mediaType := range response.Content {
			if mediaType.Schema != nil {
				w.Header().Set("Content-Type", contentType)
				w.WriteHeader(statusCode)

				data := s.generator.Generate(mediaType.Schema)
				if err := json.NewEncoder(w).Encode(data); err != nil {
					log.Printf("Error encoding response: %v", err)
				}
				return
			}
		}

		// No content defined
		w.WriteHeader(statusCode)
	}
}
