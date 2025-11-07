package main

import (
	"encoding/binary"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/savageking-io/ogbrest/packet"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// @TODO: Implement origin check
		return isOriginAllowed(r)
	},
	EnableCompression: true,
}

func isOriginAllowed(r *http.Request) bool {
	origins := AppConfig.Rest.AllowedOrigins
	if len(origins) == 0 {
		return true
	}
	origin := r.Header.Get("Origin")
	for _, allowedOrigin := range origins {
		if origin == allowedOrigin {
			return true
		}
	}
	return false
}

func NewWebSocketClient(w http.ResponseWriter, req *http.Request) (*WebSocketClient, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}

	conn, err := wsUpgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Errorf("Failed to upgrade connection: %s", err.Error())
		return nil, err
	}
	return &WebSocketClient{conn: conn}, nil
}

type WebSocketClient struct {
	conn     *websocket.Conn
	shutdown bool
}

func (c *WebSocketClient) Run() {
	log.Traceln("WebSocketClient::Run")
	defer c.conn.Close()
	for !c.shutdown {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Errorf("Failed to read message: %s", err.Error())
			return
		}

		if messageType == websocket.BinaryMessage {
			if err := c.HandleBinaryMessage(message); err != nil {
				log.Errorf("Failed to handle binary message: %s", err.Error())
			}
			continue
		}

		if messageType == websocket.TextMessage {
			if err := c.HandleTextMessage(message); err != nil {
				log.Errorf("Failed to handle text message: %s", err.Error())
			}
			continue
		}

		if messageType == websocket.CloseMessage {
			if err := c.HandleCloseMessage(message); err != nil {
				log.Errorf("Failed to handle close message: %s", err.Error())
			}
			continue
		}
	}
	return
}

func (c *WebSocketClient) HandleBinaryMessage(message []byte) error {
	log.Traceln("WebSocketClient::HandleBinaryMessage")
	// Read magic byte - if present that means connection is coming from a game or other headless client
	// Otherwise it's a web browser and we should expect json

	if len(message) > 16 && binary.BigEndian.Uint16(message[0:2]) == packet.MagicProtobuf {
		return c.HandleProtobuf(message)
	}

	return c.HandleJson(message)
}

func (c *WebSocketClient) HandleTextMessage(message []byte) error {
	log.Traceln("WebSocketClient::HandleTextMessage")
	return nil
}

func (c *WebSocketClient) HandleCloseMessage(message []byte) error {
	log.Traceln("WebSocketClient::HandleCloseMessage")
	c.shutdown = true
	return nil
}

func (c *WebSocketClient) HandleProtobuf(message []byte) error {
	log.Traceln("WebSocketClient::HandleProtobuf")
	if c.conn == nil {
		return fmt.Errorf("nil connection")
	}
	p, err := packet.Unmarshal(message)
	if err != nil {
		log.Errorf("Failed to unmarshal packet: %s", err.Error())
		return err
	}

	if p.Magic != packet.MagicProtobuf {
		log.Errorf("Trying to handle protobuf message, but Magic bit was different")
		return fmt.Errorf("bad magic byte")
	}

	return nil
}

func (c *WebSocketClient) HandleJson(message []byte) error {
	log.Traceln("WebSocketClient::HandleJson")
	return nil
}
