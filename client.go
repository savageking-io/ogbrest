package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/savageking-io/ogbrest/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

// RegisterNewRouteHandler A handle from REST to add new routes
type RegisterNewRouteHandler func(root, method, uri string, client *Client) error

// AddRouteToIgnoreListHandler A handle from REST to add path to auth ignore list
type AddRouteToIgnoreListHandler func(uri string)

type Client struct {
	Label                       string // Unique label of the service to help developers identify it
	Host                        string
	Port                        uint16
	Token                       string
	ServiceId                   uint16 // ServiceId provided by the client during the authentication step
	conn                        *grpc.ClientConn
	client                      proto.RestInterServiceClient
	registerNewRouteHandler     RegisterNewRouteHandler
	addRouteToIgnoreListHandler AddRouteToIgnoreListHandler
}

func (c *Client) Init(config *ServiceConfig, routeRegistrationHandler RegisterNewRouteHandler, addRouteToIgnoreListHandler AddRouteToIgnoreListHandler) error {
	log.Traceln("Client::Init")
	if config == nil {
		return fmt.Errorf("no service configuration provided")
	}
	log.Infof("Initializing client [%s]", config.Label)
	c.Label = config.Label
	c.Host = config.Hostname
	c.Port = config.Port
	c.Token = config.Token
	c.registerNewRouteHandler = routeRegistrationHandler
	c.addRouteToIgnoreListHandler = addRouteToIgnoreListHandler
	return nil
}

func (c *Client) Start() error {
	log.Traceln("Client::Start")
	var err error
	log.Infof("Connecing client [%s] to %s:%d", c.Label, c.Host, c.Port)
	c.conn, err = grpc.NewClient(fmt.Sprintf("%s:%d", c.Host, c.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	log.Infof("Client [%s] connected", c.Label)
	if c.authenticate() != nil {
		return fmt.Errorf("authentication failed")
	}

	log.Infof("Client [%s] authenticated. Requesting REST data", c.Label)
	if c.requestRestData() != nil {
		return fmt.Errorf("requesting REST data failed")
	}
	log.Infof("Client [%s] start sequence complete", c.Label)

	return nil
}

func (c *Client) ScheduleRestart() {
	// @TODO: Make wait time configurable
	log.Traceln("Client::ScheduleRestart")
	waitTime := time.Second * 3
	log.Infof("Scheduling restart of client [%s] in %s", c.Label, waitTime.String())
	go func() {
		startedAt := time.Now()
		for {
			time.Sleep(time.Millisecond * 100)
			if time.Since(startedAt) > waitTime {
				log.Infof("Restarting client [%s]", c.Label)
				if err := c.Start(); err != nil {
					log.Errorf("Failed to restart client [%s]: %s", c.Label, err.Error())
					c.ScheduleRestart()
					return
				}
				log.Infof("Client [%s] restarted successfully", c.Label)
				break
			}
		}
	}()
}

func (c *Client) authenticate() error {
	log.Traceln("Client::authenticate")
	if c.conn == nil {
		return fmt.Errorf("connection is not initialized")
	}
	c.client = proto.NewRestInterServiceClient(c.conn)

	log.Infof("Authenticating client [%s]", c.Label)

	authRequest := &proto.AuthenticateServiceRequest{
		Token: c.Token,
	}
	authResponse, err := c.client.AuthInterService(context.Background(), authRequest)
	if err != nil {
		log.Errorf("Authentication failed for client [%s]: %s", c.Label, err.Error())
		return err
	}
	if authResponse.Code != 0 {
		log.Errorf("Authentication failed for client [%s]. Code %d: %s", c.Label, authResponse.Code, authResponse.Error)
		return fmt.Errorf("authentication failed")
	}
	return nil
}

func (c *Client) requestRestData() error {
	log.Traceln("Client::requestRestData")
	if c.conn == nil {
		return fmt.Errorf("connection is not initialized")
	}
	if c.client == nil {
		return fmt.Errorf("client is not initialized")
	}

	log.Infof("Requesting REST data for client [%s]", c.Label)

	// @TODO: Add version support
	restRequest := &proto.RestDataRequest{}
	restResponse, err := c.client.RequestRestData(context.Background(), restRequest)
	if err != nil {
		log.Errorf("Requesting REST data failed for client [%s]: %s", c.Label, err.Error())
		return err
	}
	if restResponse.Code != 0 {
		log.Errorf("Requesting REST data failed for client [%s]. Code %d: %s", c.Label, restResponse.Code, restResponse.Error)
		return fmt.Errorf("requesting REST data failed")
	}

	for _, endpoint := range restResponse.Endpoints {
		log.Infof("Registering route %s:%s for client [%s]", endpoint.Method, endpoint.Path, c.Label)
		if err := c.registerNewRouteHandler(restResponse.Root, endpoint.Method, endpoint.Path, c); err != nil {
			log.Errorf("Registering route failed: %s", err.Error())
			return err
		}
		if endpoint.SkipAuthMiddleware {
			log.Infof("Adding route %s:/%s%s to ignore list for client [%s]", endpoint.Method, restResponse.Root, endpoint.Path, c.Label)
			c.addRouteToIgnoreListHandler(fmt.Sprintf("%s:/%s%s", endpoint.Method, restResponse.Root, endpoint.Path))
		}
	}

	return nil
}

func (c *Client) HandleRestRequest(request *proto.RestApiRequest) (*proto.RestApiResponse, error) {
	log.Traceln("Client::HandleRestRequest")
	if c.conn == nil {
		return nil, fmt.Errorf("connection is not initialized")
	}
	if c.client == nil {
		return nil, fmt.Errorf("client is not initialized")
	}
	if request == nil {
		return nil, fmt.Errorf("request is not initialized")
	}

	log.Debugf("Handling REST request %s:%s for client [%s]", request.Method, request.Uri, c.Label)

	restResponse, err := c.client.NewRestRequest(context.Background(), request)
	if err != nil {
		if errors.Is(err, grpc.ErrServerStopped) {
			c.ScheduleRestart()
			return &proto.RestApiResponse{
				Code: 503,
			}, err
		}
		log.Warnf("Handling REST request failed for client [%s]: %s", c.Label, err.Error())
		return restResponse, err
	}

	if restResponse == nil {
		log.Warnf("Received empty response from [%s]", c.Label)
	}

	return restResponse, nil
}
