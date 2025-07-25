package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/savageking-io/ogbrest/proto"
	"github.com/savageking-io/ogbrest/user_client"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

type Route struct {
	Method string
	Root   string
	Uri    string
}

type REST struct {
	Hostname               string
	Port                   uint16
	AllowedOrigins         []string
	mux                    *chi.Mux
	RoutesExcludedFromAuth []string
	UserService            *user_client.UserClient
}

func (r *REST) Init(inConfig *RestConfig, user *user_client.UserClient) error {
	if inConfig == nil {
		return fmt.Errorf("no configuration")
	}

	if user == nil {
		return fmt.Errorf("no user service provided")
	}

	r.UserService = user

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

func (r *REST) AddToAuthIgnoreList(uri string) {
	r.RoutesExcludedFromAuth = append(r.RoutesExcludedFromAuth, uri)
}

func (r *REST) JWTMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// @TODO: Fix this
			log.Tracef("[JWTMiddleware] Request: %s %s", req.Method, req.URL.Path)
			for _, path := range r.RoutesExcludedFromAuth {
				if req.URL.Path == path || strings.HasPrefix(req.URL.Path, path) {
					next.ServeHTTP(w, req)
					return
				}
			}

			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			if r.UserService == nil {
				http.Error(w, "User service is not initialized", http.StatusServiceUnavailable)
				return
			}

			isValid, userId, err := r.UserService.ValidateToken(context.Background(), tokenString)
			if err != nil {
				log.Errorf("Failed to validate token: %s", err.Error())
				http.Error(w, "Failed to validate token", http.StatusUnauthorized)
				return
			}

			if !isValid {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			log.Tracef("[JWTMiddleware] Request handled")
			ctx := context.WithValue(req.Context(), "user_id", userId)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
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
		request.Uri = uri

		response, err := client.HandleRestRequest(request)
		if err != nil {
			log.Errorf("Failed to handle REST request: %s", err.Error())

			// In a normal scenario a service must provide http code that should be returned to the client
			if response != nil && response.HttpCode != 0 {
				w.WriteHeader(int(response.HttpCode))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

			if response != nil && response.Code != 0 {
				// We have an internal error code provided. Return it to the client as json
				ret := make(map[string]interface{})
				ret["code"] = response.Code
				ret["error"] = response.Error
				ret["date"] = time.Now().String()
				responseBody, _ := json.Marshal(ret)
				_, _ = w.Write(responseBody)
			}

			return
		}

		// This is not good - blame the service for bad implementation
		if response == nil {
			log.Errorf("Failed to handle REST request: no response")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, header := range response.Headers {
			w.Header().Set(header.Key, header.Value)
		}

		w.WriteHeader(int(response.HttpCode))
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
		Method:  req.Method,
		Headers: headers,
		Body:    bodyString,
		Source:  req.RemoteAddr,
		Form:    formData,
	}
}
