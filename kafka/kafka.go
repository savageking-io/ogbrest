package kafka

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers      []string `yaml:"brokers"`       // list of brokers: ["localhost:9092"]
	Topic        string   `yaml:"topic"`         // topic to write events to
	ClientID     string   `yaml:"client_id"`     // optional client id
	Compression  string   `yaml:"compression"`   // none|gzip|snappy|lz4|zstd
	RequiredAcks int      `yaml:"required_acks"` // -1(all), 1(leader), 0(no ack)
}

type Publisher struct {
	writer  *kafka.Writer
	enabled bool
}

type RequestSchema struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   string            `json:"query"`
	Headers map[string]string `json:"headers"`
	Source  string            `json:"source"`
}

var eventPublisher = &Publisher{}

func parseCompression(compression string) kafka.Compression {
	switch strings.ToLower(compression) {
	case "gzip":
		return kafka.Gzip
	case "snappy":
		return kafka.Snappy
	case "lz4":
		return kafka.Lz4
	case "zstd":
		return kafka.Zstd
	default:
		return kafka.Snappy // reasonable default
	}
}

func (p *Publisher) Init(cfg Config) error {
	// Basic validation
	if len(cfg.Brokers) == 0 || strings.TrimSpace(cfg.Topic) == "" {
		log.Warn("Kafka not configured: events will not be published")
		p.enabled = false
		return nil
	}

	balance := &kafka.LeastBytes{}
	compression := parseCompression(cfg.Compression)

	acks := kafka.RequireAll
	switch cfg.RequiredAcks {
	case 0:
		acks = kafka.RequireNone
	case 1:
		acks = kafka.RequireOne
	case -1:
		acks = kafka.RequireAll
	}

	p.writer = &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  cfg.Topic,
		Balancer:               balance,
		AllowAutoTopicCreation: true,
		Compression:            compression,
		RequiredAcks:           acks,
	}
	p.enabled = true
	if cfg.ClientID != "" {
		p.writer.Transport = &kafka.Transport{ClientID: cfg.ClientID}
	}
	log.Infof("Kafka publisher initialized. Brokers=%v Topic=%s", cfg.Brokers, cfg.Topic)
	return nil
}

func (p *Publisher) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// Publish writes a single message to Kafka. key/value are raw bytes.
func (p *Publisher) Publish(ctx context.Context, key, value []byte) error {
	if !p.enabled {
		return nil
	}
	log.Traceln("Kafka::Publisher::Publish")
	msg := kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}
	return p.writer.WriteMessages(ctx, msg)
}

func (p *Publisher) LogRequest(req *http.Request) {
	log.Traceln("Kafka::Publisher::LogRequest")
	if !p.enabled {
		log.Debugf("Skipping request logging: Kafka not enabled")
		return
	}
	go p.logRequestInternal(req)
}

func (p *Publisher) logRequestInternal(req *http.Request) {
	log.Traceln("Kafka::Publisher::logRequestInternal")
	r := &RequestSchema{
		Method:  req.Method,
		Path:    req.RequestURI,
		Query:   req.URL.RawQuery,
		Headers: make(map[string]string),
		Source:  req.RemoteAddr,
	}
	for k, v := range req.Header {
		r.Headers[k] = v[0]
	}

	data, err := json.Marshal(r)
	if err != nil {
		log.Errorf("Failed to marshal request schema: %s", err.Error())
		return
	}

	if err := p.Publish(context.Background(), []byte("req"), data); err != nil {
		log.Errorf("Failed to publish request to Kafka: %s", err.Error())
		return
	}
}
