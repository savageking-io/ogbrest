package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/savageking-io/ogbrest/kafka"
	"github.com/savageking-io/ogbrest/proto"
	user_client "github.com/savageking-io/ogbuser/client"
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
	UserService            *user_client.Client
	kafka                  *kafka.Publisher
}

func (r *REST) Init(inConfig *RestConfig, kafkaConfig kafka.Config, user *user_client.Client) error {
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
	// Apply JWT middleware globally; it will skip paths present in RoutesExcludedFromAuth
	r.mux.Use(r.JWTMiddleware())

	// Default exclusions
	r.AddToAuthIgnoreList("/status")

	r.kafka = new(kafka.Publisher)
	if err := r.kafka.Init(kafkaConfig); err != nil {
		return err
	}

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
			requestPath := req.URL.Path
			requestKey := fmt.Sprintf("%s:%s", req.Method, requestPath)
			for _, ex := range r.RoutesExcludedFromAuth {
				log.Warnf("[JWTMiddleware] Checking path %s against %s", ex, requestPath)
				// Exact or prefix path match (e.g., "/status" or "/api/public")
				if requestPath == ex || strings.HasPrefix(requestPath, ex) {
					next.ServeHTTP(w, req)
					return
				}
				// Support entries like METHOD:PATH coming from services, where PATH may be without root
				// Match exact METHOD:FULL_PATH or suffix METHOD:ENDPOINT_PATH
				if strings.Contains(ex, ":") {
					if requestKey == ex || strings.HasSuffix(requestKey, ex) {
						next.ServeHTTP(w, req)
						return
					}
				} else {
					// Also allow suffix match on path (ignore unknown root prefix)
					if strings.HasSuffix(requestPath, ex) {
						next.ServeHTTP(w, req)
						return
					}
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
		r.kafka.LogRequest(req)
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

			if response != nil {
				log.Debugf("Response from failed RPC call: %+v", response)
			}

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
			log.Tracef("Writing header %s: %s", header.Key, header.Value)
			w.Header().Set(header.Key, header.Value)
		}

		if response.HttpCode != 0 {
			w.WriteHeader(int(response.HttpCode))
		} else {
			log.Warnf("No HTTP code provided for response. Using 200. Check service implementation")
		}

		if response.Body == "" && response.Error != "" {
			response.Body = "{'error': '" + response.Error + "'}"
		}

		log.Tracef("Writing response body: %s", response.Body)
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
