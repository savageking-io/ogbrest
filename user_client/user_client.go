package user_client

import (
	"context"
	"errors"
	"fmt"
	"github.com/savageking-io/ogbuser/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sync"
	"time"
)

// UserClient connects to user microservice for JWT operations
type UserClient struct {
	conn      *grpc.ClientConn
	client    proto.UserServiceClient
	hostname  string
	port      uint16
	mutex     sync.Mutex
	ErrorChan chan error
}

func NewUserClient() *UserClient {
	return &UserClient{}
}

func (c *UserClient) Init(hostname string, port uint16) error {
	if hostname == "" {
		return fmt.Errorf("hostname is not provided")
	}
	if port == 0 {
		return fmt.Errorf("port is not provided")
	}
	c.hostname = hostname
	c.port = port
	c.ErrorChan = make(chan error)
	return nil
}

func (c *UserClient) Run() error {
	log.Infof("Connecting to user microservice at %s:%d", c.hostname, c.port)
	var err error

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn != nil {
		return nil
	}

	c.conn, err = grpc.NewClient(fmt.Sprintf("%s:%d", c.hostname, c.port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	c.client = proto.NewUserServiceClient(c.conn)

	lastPing := time.Unix(0, 0)
	for {
		if time.Since(lastPing) > time.Second*5 {
			if err := c.Ping(); err != nil {
				log.Errorf("Ping to user microservice failed: %s", err.Error())
				return err
			}
			lastPing = time.Now()
		}
	}
}

func (c *UserClient) Stop() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn == nil {
		return fmt.Errorf("already closed")
	}

	if err := c.conn.Close(); err != nil {
		return err
	}
	c.conn = nil
	c.client = nil
	return nil
}

func (c *UserClient) ValidateToken(ctx context.Context, token string) (bool, int32, error) {
	if c.conn == nil {
		return false, -1, fmt.Errorf("connection is not initialized")
	}
	if c.client == nil {
		return false, -1, fmt.Errorf("client is not initialized")
	}

	log.Debugf("Validating token %s", token)

	result, err := c.client.ValidateToken(ctx, &proto.ValidateTokenRequest{Token: token})
	if err != nil {
		if errors.Is(err, grpc.ErrServerStopped) {
			go func() {
				// @TODO: We should not call start all the time, but after some period if reconnect failed
				log.Errorf("Connection to user microservice lost. Reconnecting...")
				if err := c.Run(); err != nil {
					log.Errorf("Connection to user microservice failed: %s", err.Error())
				}
			}()
		}
		return false, -1, err
	}
	if result.Code != 0 {
		return false, -1, fmt.Errorf("validation failed with code %d: %s", result.Code, result.Error)
	}

	log.Debugf("Token %s validation result: %t", token, result.IsValid)
	return result.IsValid, result.UserId, nil
}

// Ping will send a ping message to user microservice
// If service is shutdown it will initiate restart
func (c *UserClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	resp, err := c.client.Ping(ctx, &proto.PingMessage{SentAt: timestamppb.New(time.Now())})
	if err != nil {
		log.Errorf("Ping to user microservice failed: %s", err.Error())
		return err
	}
	sentAt := resp.SentAt.AsTime()
	repliedAt := resp.RepliedAt.AsTime()
	diff := repliedAt.Sub(sentAt)
	log.Debugf("Ping to user microservice replied in %s", diff.String())
	return nil
}
