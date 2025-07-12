package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"net/http"
	"time"
)

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

	return nil
}

func (r *REST) Start() error {

	r.mux = chi.NewMux()
	r.mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   r.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

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
