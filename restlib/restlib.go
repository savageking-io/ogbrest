// Package restlib provides functions for OGB microservices to authenticate and communicate with REST microservice
package restlib

import (
	"fmt"
	"github.com/savageking-io/ogbrest/proto"
)

// RestServiceConfig - Configuration for the REST service that should be defined inside a microservice
type RestServiceConfig struct {
	isInitialized   bool
	isAuthenticated bool
	AuthToken       string                `yaml:"auth_token"`
	EndpointRoot    string                `yaml:"endpoint_root"`
	Endpoints       []RestServiceEndpoint `yaml:"endpoints"`
}

// RestServiceEndpoint defines endpoints that will be available to users and handled/forwarded by REST API
type RestServiceEndpoint struct {
	Endpoint   string `yaml:"endpoint"` // Endpoint should be a top-level path
	Method     string `yaml:"method"`
	IsDisabled bool   `yaml:"is_disabled"`
	// @TODO: Consider having a boolean whether auth required or not
	// @TODO: Consider having definitions for extra headers
}

var serviceConfig RestServiceConfig

// SetServiceConfig must be called after configuration loading is done
func SetServiceConfig(config RestServiceConfig) {
	serviceConfig = config
	serviceConfig.isInitialized = true
}

// HandleIncomingAuthRequest will compare the incoming token to the one defined in RestServiceConfig
func HandleIncomingAuthRequest(packet *proto.AuthenticateServiceRequest) (*proto.AuthenticateServiceResponse, error) {
	if !serviceConfig.isInitialized {
		// @TODO: Should we panic?
		return nil, fmt.Errorf("service config not initialized")
	}

	if packet == nil {
		return nil, fmt.Errorf("nil packet")
	}

	if packet.Token == "" {
		return nil, fmt.Errorf("empty token")
	}

	if packet.Token == serviceConfig.AuthToken {
		serviceConfig.isAuthenticated = true
		return &proto.AuthenticateServiceResponse{
			Code: 200,
		}, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// HandleIncomingEndpointsRequest will return endpoints defined in configuration
func HandleIncomingEndpointsRequest(packet *proto.RestDataRequest) (*proto.RestDataResponse, error) {
	if !serviceConfig.isInitialized {
		return nil, fmt.Errorf("service config not initialized")
	}

	if !serviceConfig.isAuthenticated {
		return nil, fmt.Errorf("not authenticated")
	}

	// @TODO: Handle version
	response := &proto.RestDataResponse{
		Code:         200,
		EndpointsNum: int32(len(serviceConfig.Endpoints)),
		Root:         serviceConfig.EndpointRoot,
	}
	return response, nil
}
