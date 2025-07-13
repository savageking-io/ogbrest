// Package restlib provides functions for OGB microservices to authenticate and communicate with REST microservice
package restlib

import (
	"context"
	"fmt"
	restproto "github.com/savageking-io/ogbrest/proto"
	"google.golang.org/grpc"
	"net"
)

type RestRequestHandler func(ctx context.Context, in *restproto.RestApiRequest) (*restproto.RestApiResponse, error)

type RestInterServiceConfig struct {
	Hostname  string                     `yaml:"hostname"`
	Port      uint16                     `yaml:"port"`
	Token     string                     `yaml:"token"`
	Root      string                     `yaml:"root"`
	Endpoints []RestInterServiceEndpoint `yaml:"endpoints"`
}

type RestInterServiceEndpoint struct {
	Path   string `yaml:"path"`
	Method string `yaml:"method"`
}

type RestInterServiceServer struct {
	restproto.UnimplementedRestInterServiceServer
	config          RestInterServiceConfig
	isAuthenticated bool
	RequestChan     chan *restproto.RestApiRequest
	handlers        map[string]RestRequestHandler
}

func NewRestInterServiceServer(config RestInterServiceConfig) *RestInterServiceServer {
	return &RestInterServiceServer{
		config: config,
	}
}

func (s *RestInterServiceServer) GetConfig() RestInterServiceConfig {
	return s.config
}

func (s *RestInterServiceServer) IsAuthenticated() bool {
	return s.isAuthenticated
}

func (s *RestInterServiceServer) Init() error {
	s.RequestChan = make(chan *restproto.RestApiRequest, 100)
	s.handlers = make(map[string]RestRequestHandler)
	return nil
}

func (s *RestInterServiceServer) Start() error {
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

func (s *RestInterServiceServer) RegisterHandler(uri, method string, handler RestRequestHandler) error {
	if s.handlers == nil {
		return fmt.Errorf("handlers are not initialized")
	}
	requestDefinition := fmt.Sprintf("%s:%s", method, uri)
	if _, ok := s.handlers[requestDefinition]; ok {
		return fmt.Errorf("handler for %s already registered", requestDefinition)
	}
	s.handlers[requestDefinition] = handler
	return nil
}

func (s *RestInterServiceServer) IsHandlerRegistered(uri, method string) bool {
	requestDefinition := fmt.Sprintf("%s:%s", method, uri)
	_, ok := s.handlers[requestDefinition]
	return ok
}

func (s *RestInterServiceServer) UnregisterHandler(uri, method string) error {
	requestDefinition := fmt.Sprintf("%s:%s", method, uri)
	if _, ok := s.handlers[requestDefinition]; !ok {
		return fmt.Errorf("handler for %s is not registered", requestDefinition)
	}
	delete(s.handlers, requestDefinition)
	return nil
}

func (s *RestInterServiceServer) UnregisterAllHandlers() error {
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
			Path:   endpoint.Path,
			Method: endpoint.Method,
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
