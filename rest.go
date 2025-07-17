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
	fullUri := fmt.Sprintf("%s%s", sanitizeRoot(root), sanitizeUri(uri))

	r.mux.MethodFunc(method, fullUri, func(w http.ResponseWriter, req *http.Request) {
		request := r.httpRequestToProto(req)
		if request == nil {
			log.Errorf("Failed to convert HTTP request to proto")
			w.WriteHeader(http.StatusBadRequest)
			return
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

		if response.Code == 0 {
			// Another service should return valid HTTP code
			log.Warnf("Invalid response code for request: %s %s", method, uri)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(int(response.Code))
		_, _ = w.Write([]byte(response.Body))
		return
	})

	return nil
}

func (r *REST) httpRequestToProto(req *http.Request) *proto.RestApiRequest {
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
		return nil
	}
	var bodyString string
	if len(body) > 0 {
		bodyString = string(body)
	}

	var formData []*proto.RestApiFormData
	if req.Method == "POST" {
		err = req.ParseForm()
		if err != nil {
			log.Errorf("Failed to parse request body: %s", err.Error())
			return nil
		}
		for key, values := range req.Form {
			form := &proto.RestApiFormData{
				Key: key,
			}
			for _, value := range values {
				form.Value = append(form.Value, value)
			}
			formData = append(formData, form)
		}
	}

	return &proto.RestApiRequest{
		Uri:     req.URL.Path,
		Method:  req.Method,
		Headers: headers,
		Body:    bodyString,
		Source:  req.RemoteAddr,
		Form:    formData,
	}
}
