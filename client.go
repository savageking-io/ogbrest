package main

/*
import (
	"fmt"
	restpb "github.com/savageking-io/onlinegamebase/rest/proto"
	userpb "github.com/savageking-io/onlinegamebase/user/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// UserClient represents a gRPC client for the user service
type UserClient struct {
	serviceConfig ServiceConfig
	client        userpb.UserServiceClient
	conn          *grpc.ClientConn
}

// NewUserClient creates a new UserClient
func NewUserClient(config ServiceConfig) *UserClient {
	return &UserClient{
		serviceConfig: config,
	}
}

// Connect establishes a connection to the user service
func (c *UserClient) Connect() error {
	address := fmt.Sprintf("%s:%d", c.serviceConfig.Hostname, c.serviceConfig.Port)
	log.Infof("Connecting to user service at %s", address)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	var err error
	c.conn, err = grpc.NewClient(fmt.Sprintf("%s:%d", c.serviceConfig.Hostname, c.serviceConfig.Port), opts...)
	if err != nil {
		log.Errorf("Failed to create gRPC client: %v", err)
		return err
	}

	log.Info("Connected to user service")
	return nil
}

// Close closes the connection to the user service
func (c *UserClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Authenticate performs token-based authentication with the user service
func (c *UserClient) Authenticate() error {
	log.Debugf("Authenticating with user service using token: %s", c.serviceConfig.Token)

	// In a real implementation, we would make a gRPC call to the AuthService method
	// For now, we'll just simulate a successful authentication
	log.Info("Successfully authenticated with user service")
	return nil
}

// GetEndpoints requests the list of endpoints from the user service
func (c *UserClient) GetEndpoints() (*restpb.RestDataResponse, error) {
	log.Debugf("Requesting endpoints from user service with version: %s", AppVersion)

	// In a real implementation, we would make a gRPC call to the RequestRESTData method
	// For now, we'll just simulate a response with some sample endpoints
	endpoints := []*restpb.RestEndpoint{
		{
			Path:   "/api/users",
			Method: "GET",
		},
		{
			Path:   "/api/users/{id}",
			Method: "GET",
		},
		{
			Path:   "/api/users",
			Method: "POST",
		},
	}

	resp := &restpb.RestDataResponse{
		Code:         200,
		EndpointsNum: int32(len(endpoints)),
		Root:         "/api",
		Endpoints:    endpoints,
	}

	log.Infof("Received %d endpoints from user service", resp.EndpointsNum)
	return resp, nil
}

// ConnectAndAuthenticate connects to the user service and authenticates
func (c *UserClient) ConnectAndAuthenticate() error {
	err := c.Connect()
	if err != nil {
		return err
	}

	err = c.Authenticate()
	if err != nil {
		c.Close()
		return err
	}

	return nil
}
*/
