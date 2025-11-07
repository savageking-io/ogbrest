// Package restlib provides functions for OGB microservices to authenticate and communicate with REST microservice
//
// Each service microservice will act as a gRPC server that ogbrest will connect to as a client. Therefore it's
// necessary to tell ogbrest which services it should connect to using a configuration file.
package restlib

import (
	"context"
	"fmt"
	restproto "github.com/savageking-io/ogbrest/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

// RestRequestHandler is a callback function called for appropriate REST requests
type RestRequestHandler func(ctx context.Context, in *restproto.RestApiRequest) (*restproto.RestApiResponse, error)

// RestInterServiceConfig is a main configuration for the microservice that will expect connections from ogbrest
type RestInterServiceConfig struct {
	Hostname  string                     `yaml:"hostname"`  // Hostname to connect to
	Port      uint16                     `yaml:"port"`      // Port to connect to
	Token     string                     `yaml:"token"`     // Token is a unique token for the service. Keep it secret
	Root      string                     `yaml:"root"`      // Root of the query string. All the requests coming to /root/ will be redirected to this microservice
	Endpoints []RestInterServiceEndpoint `yaml:"endpoints"` // Endpoints list all the endpoints available
}

// RestInterServiceEndpoint defines REST API endpoint
type RestInterServiceEndpoint struct {
	Path               string `yaml:"path"`                 // Path will be appended to RestInterServiceConfig.Root
	Method             string `yaml:"method"`               // Method can be GET, POST, DELETE, PUT, UPDATE or any other valid method
	SkipAuthMiddleware bool   `yaml:"skip_auth_middleware"` // SkipAuthMiddleware will not check user's Auth token for this endpoint
}

// RestInterServiceServer
type RestInterServiceServer struct {
	restproto.UnimplementedRestInterServiceServer
	config          RestInterServiceConfig
	isAuthenticated bool
	handlers        map[string]RestRequestHandler
	RequestChan     chan *restproto.RestApiRequest
}

// NewRestInterServiceServer will create new RestInterServiceServer with the provided confiration
func NewRestInterServiceServer(config RestInterServiceConfig) *RestInterServiceServer {
	log.Traceln("RestLib::NewRestInterServiceServer")
	return &RestInterServiceServer{
		config: config,
	}
}

func (s *RestInterServiceServer) GetConfig() RestInterServiceConfig {
	return s.config
}

// IsAuthenticated will return true if ogbrest passed authentication successfully
func (s *RestInterServiceServer) IsAuthenticated() bool {
	return s.isAuthenticated
}

func (s *RestInterServiceServer) Init() error {
	log.Traceln("RestLib::Init")
	s.RequestChan = make(chan *restproto.RestApiRequest, 100)
	s.handlers = make(map[string]RestRequestHandler)
	return nil
}

// Start will create TCP Listener and enable gRPC server
func (s *RestInterServiceServer) Start() error {
	log.Traceln("RestLib::Start")
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.Hostname, s.config.Port))
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	restproto.RegisterRestInterServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

// RegisterHandler will add new URL to the rest service
func (s *RestInterServiceServer) RegisterHandler(uri, method string, handler RestRequestHandler, skipAuth bool) error {
	log.Traceln("RestLib::RegisterHandler")
	if s.handlers == nil {
		return fmt.Errorf("handlers are not initialized")
	}
	requestDefinition := fmt.Sprintf("%s:%s", method, uri)
	if _, ok := s.handlers[requestDefinition]; ok {
		return fmt.Errorf("handler for %s already registered", requestDefinition)
	}
	s.handlers[requestDefinition] = handler
	if skipAuth {

	}
	return nil
}

func (s *RestInterServiceServer) IsHandlerRegistered(uri, method string) bool {
	requestDefinition := fmt.Sprintf("%s:%s", method, uri)
	_, ok := s.handlers[requestDefinition]
	return ok
}

func (s *RestInterServiceServer) UnregisterHandler(uri, method string) error {
	log.Traceln("RestLib::UnregisterHandler")
	requestDefinition := fmt.Sprintf("%s:%s", method, uri)
	if _, ok := s.handlers[requestDefinition]; !ok {
		return fmt.Errorf("handler for %s is not registered", requestDefinition)
	}
	delete(s.handlers, requestDefinition)
	return nil
}

func (s *RestInterServiceServer) UnregisterAllHandlers() error {
	log.Traceln("RestLib::UnregisterAllHandlers")
	s.handlers = make(map[string]RestRequestHandler)
	return nil
}

func (s *RestInterServiceServer) GetRegisteredHandlerKeys() []string {
	keys := make([]string, len(s.handlers))
	i := 0
	for k := range s.handlers {
		keys[i] = k
		i++
	}
	return keys
}

func (s *RestInterServiceServer) AuthInterService(ctx context.Context, in *restproto.AuthenticateServiceRequest) (*restproto.AuthenticateServiceResponse, error) {
	log.Traceln("RestLib::AuthInterService")
	if s.config.Token == "" {
		return nil, fmt.Errorf("token is not set")
	}
	if in.Token != s.config.Token {
		return &restproto.AuthenticateServiceResponse{
			Code:  1,
			Error: "invalid token",
		}, nil
	}
	s.isAuthenticated = true
	return &restproto.AuthenticateServiceResponse{
		Code: 0,
	}, nil
}

func (s *RestInterServiceServer) RequestRestData(ctx context.Context, in *restproto.RestDataRequest) (*restproto.RestDataDefinition, error) {
	log.Traceln("RestLib::RequestRestData")
	if s.config.Token == "" {
		return nil, fmt.Errorf("token is not set")
	}

	if !s.isAuthenticated {
		return &restproto.RestDataDefinition{
			Code:  1,
			Error: "not authenticated",
		}, fmt.Errorf("not authenticated")
	}

	endpoints := make([]*restproto.RestEndpoint, len(s.config.Endpoints))
	for i, endpoint := range s.config.Endpoints {
		endpoints[i] = &restproto.RestEndpoint{
			Path:               endpoint.Path,
			Method:             endpoint.Method,
			SkipAuthMiddleware: endpoint.SkipAuthMiddleware,
		}
	}

	return &restproto.RestDataDefinition{
		Code:         0,
		Root:         s.config.Root,
		Endpoints:    endpoints,
		EndpointsNum: int32(len(s.config.Endpoints)),
	}, nil
}

func (s *RestInterServiceServer) NewRestRequest(ctx context.Context, in *restproto.RestApiRequest) (*restproto.RestApiResponse, error) {
	log.Traceln("RestLib::NewRestRequest")
	if s.config.Token == "" {
		return nil, fmt.Errorf("token is not set")
	}
	if !s.isAuthenticated {
		return nil, fmt.Errorf("not authenticated")
	}

	requestDefinition := fmt.Sprintf("%s:%s", in.Method, in.Uri)
	handler, ok := s.handlers[requestDefinition]
	if !ok {
		return &restproto.RestApiResponse{
			Code: 404,
		}, fmt.Errorf("handler for %s is not registered", requestDefinition)
	}
	return handler(ctx, in)
}
