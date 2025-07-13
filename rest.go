package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/savageking-io/ogbrest/proto"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

type Route struct {
	Method string
	Root   string
	Uri    string
}

type REST struct {
	Hostname       string
	Port           uint16
	AllowedOrigins []string
	mux            *chi.Mux
}

func (r *REST) Init(inConfig *RestConfig) error {
	if inConfig == nil {
		return fmt.Errorf("no configuration")
	}

	r.Hostname = inConfig.Hostname
	r.Port = inConfig.Port
	r.AllowedOrigins = inConfig.AllowedOrigins

	r.mux = chi.NewMux()
	r.mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   r.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	return nil
}

func (r *REST) Start() error {
	r.mux.Get("/", func(w http.ResponseWriter, req *http.Request) {
		// For default empty route return 404
		w.WriteHeader(http.StatusNotFound)
	})
	r.mux.Get("/status", r.HandleStatusRequest)

	return http.ListenAndServe(fmt.Sprintf("%s:%d", r.Hostname, r.Port), r.mux)
}

func (r *REST) HandleStatusRequest(w http.ResponseWriter, req *http.Request) {
	data := make(map[string]interface{})
	data["code"] = 0
	data["date"] = time.Now().String()

	response, _ := json.Marshal(data)
	_, err := w.Write(response)
	if err != nil {
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (r *REST) RegisterNewRoute(root, method, uri string, client *Client) error {
	root = sanitizeRoot(root)
	uri = sanitizeUri(uri)
	fullUri := fmt.Sprintf("%s%s", root, uri)

	r.mux.MethodFunc(method, fullUri, func(w http.ResponseWriter, req *http.Request) {
		var headers []*proto.RestHeader
		for k, v := range req.Header {
			headers = append(headers, &proto.RestHeader{
				Key:   k,
				Value: v[0],
			})
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Errorf("Failed to read request body: %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var bodyString string
		if len(body) > 0 {
			bodyString = string(body)
		}

		request := &proto.RestApiRequest{
			Method:  method,
			Uri:     uri,
			Headers: headers,
			Body:    bodyString,
			Source:  req.RemoteAddr,
		}
		response, err := client.HandleRestRequest(request)
		if err != nil {
			log.Errorf("Failed to handle REST request: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if response == nil {
			log.Errorf("Failed to handle REST request: no response")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, header := range response.Headers {
			w.Header().Set(header.Key, header.Value)
		}

		_, _ = w.Write([]byte(response.Body))
		w.WriteHeader(int(response.Code))
		return
	})

	return nil
}
