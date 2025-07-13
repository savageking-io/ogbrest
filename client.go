package main

import (
	"context"
	"fmt"
	"github.com/savageking-io/ogbrest/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RegisterNewRouteHandler func(root, method, uri string, client *Client) error

type Client struct {
	Label                   string
	Host                    string
	Port                    uint16
	Token                   string
	conn                    *grpc.ClientConn
	client                  proto.RestInterServiceClient
	registerNewRouteHandler RegisterNewRouteHandler
}

func (c *Client) Init(config *ServiceConfig, routeRegistrationHandler RegisterNewRouteHandler) error {
	if config == nil {
		return fmt.Errorf("no service configuration provided")
	}
	log.Infof("Initializing client %s", config.Label)
	c.Label = config.Label
	c.Host = config.Hostname
	c.Port = config.Port
	c.Token = config.Token
	c.registerNewRouteHandler = routeRegistrationHandler
	return nil
}

func (c *Client) Start() error {
	var err error
	c.conn, err = grpc.NewClient(fmt.Sprintf("%s:%d", c.Host, c.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	if c.authenticate() != nil {
		return fmt.Errorf("authentication failed")
	}

	if c.requestRestData() != nil {
		return fmt.Errorf("requesting REST data failed")
	}

	return nil
}

func (c *Client) authenticate() error {
	if c.conn == nil {
		return fmt.Errorf("connection is not initialized")
	}
	c.client = proto.NewRestInterServiceClient(c.conn)

	log.Infof("Authenticating client %s", c.Label)

	authRequest := &proto.AuthenticateServiceRequest{
		Token: c.Token,
	}
	authResponse, err := c.client.AuthInterService(context.Background(), authRequest)
	if err != nil {
		log.Errorf("Authentication failed: %s", err.Error())
		return err
	}
	if authResponse.Code != 0 {
		log.Errorf("Authentication failed with code %d: %s", authResponse.Code, authResponse.Error)
		return fmt.Errorf("authentication failed")
	}
	return nil
}

func (c *Client) requestRestData() error {
	if c.conn == nil {
		return fmt.Errorf("connection is not initialized")
	}
	if c.client == nil {
		return fmt.Errorf("client is not initialized")
	}

	log.Infof("Requesting REST data for client %s", c.Label)

	// @TODO: Add version support
	restRequest := &proto.RestDataRequest{}
	restResponse, err := c.client.RequestRestData(context.Background(), restRequest)
	if err != nil {
		log.Errorf("Requesting REST data failed: %s", err.Error())
		return err
	}
	if restResponse.Code != 0 {
		log.Errorf("Requesting REST data failed with code %d: %s", restResponse.Code, restResponse.Error)
		return fmt.Errorf("requesting REST data failed")
	}

	for _, endpoint := range restResponse.Endpoints {
		log.Infof("Registering route %s:%s for client %s", endpoint.Method, endpoint.Path, c.Label)
		if err := c.registerNewRouteHandler(restResponse.Root, endpoint.Method, endpoint.Path, c); err != nil {
			log.Errorf("Registering route failed: %s", err.Error())
			return err
		}
	}

	return nil
}

func (c *Client) HandleRestRequest(request *proto.RestApiRequest) (*proto.RestApiResponse, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("connection is not initialized")
	}
	if c.client == nil {
		return nil, fmt.Errorf("client is not initialized")
	}

	log.Infof("Handling REST request %s:%s for client %s", request.Method, request.Uri, c.Label)

	restResponse, err := c.client.NewRestRequest(context.Background(), request)
	if err != nil {
		log.Errorf("Handling REST request failed: %s", err.Error())
		return nil, err
	}

	return restResponse, nil
}
