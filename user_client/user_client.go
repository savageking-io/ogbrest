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
	conn         *grpc.ClientConn
	client       proto.UserServiceClient
	hostname     string
	port         uint16
	mutex        sync.Mutex
	isRestarting bool
}

func (c *UserClient) Init(hostname string, port uint16) error {
	c.hostname = hostname
	c.port = port
	return nil
}

func (c *UserClient) Start() error {
	log.Infof("Connecting to user microservice at %s:%d", c.hostname, c.port)
	var err error

	if c.isRestarting {
		return nil
	}

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

	resp, err := c.client.Ping(context.Background(), &proto.PingMessage{SentAt: timestamppb.New(time.Now())})
	if err != nil {
		log.Errorf("Ping to user microservice failed: %s", err.Error())
		return err
	}

	sentAt := resp.SentAt.AsTime()
	repliedAt := resp.RepliedAt.AsTime()
	diff := repliedAt.Sub(sentAt)
	log.Infof("Ping to user microservice replied in %s", diff.String())

	return err
}

// Restart should be called only from the main thread (main.go, service.go)
func (c *UserClient) Restart() {
	if c.isRestarting {
		return
	}
	c.isRestarting = true
	waitTime := time.Millisecond * 1000
	log.Infof("Scheduling restart of user client in %s", waitTime.String())
	go func() {
		startedAt := time.Now()
		for {
			time.Sleep(time.Millisecond * 100)
			if time.Since(startedAt) > waitTime {
				c.isRestarting = false
				if err := c.Start(); err != nil {
					c.Restart()
					return
				}
				break
			}
		}
	}()
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
				if err := c.Start(); err != nil {
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
