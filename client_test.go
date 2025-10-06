package main

import (
	"github.com/savageking-io/ogbrest/proto"
	"google.golang.org/grpc"
	"reflect"
	"testing"
)

func TestClient_HandleRestRequest(t *testing.T) {
	type fields struct {
		Label                       string
		Host                        string
		Port                        uint16
		Token                       string
		conn                        *grpc.ClientConn
		client                      proto.RestInterServiceClient
		registerNewRouteHandler     RegisterNewRouteHandler
		addRouteToIgnoreListHandler AddRouteToIgnoreListHandler
	}
	type args struct {
		request *proto.RestApiRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *proto.RestApiResponse
		wantErr bool
	}{
		{"Nil conn", fields{}, args{}, nil, true},
		{"Nil client", fields{conn: &grpc.ClientConn{}}, args{}, nil, true},
		{"Nil request", fields{conn: &grpc.ClientConn{}, client: proto.NewRestInterServiceClient(&grpc.ClientConn{})}, args{}, nil, true},
		{"Empty request", fields{conn: &grpc.ClientConn{}, client: proto.NewRestInterServiceClient(&grpc.ClientConn{})}, args{request: &proto.RestApiRequest{}}, nil, true},
		{"Server stopped", fields{conn: &grpc.ClientConn{}, client: proto.NewRestInterServiceClient(&grpc.ClientConn{})}, args{request: &proto.RestApiRequest{}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Label:                       tt.fields.Label,
				Host:                        tt.fields.Host,
				Port:                        tt.fields.Port,
				Token:                       tt.fields.Token,
				conn:                        tt.fields.conn,
				client:                      tt.fields.client,
				registerNewRouteHandler:     tt.fields.registerNewRouteHandler,
				addRouteToIgnoreListHandler: tt.fields.addRouteToIgnoreListHandler,
			}
			got, err := c.HandleRestRequest(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleRestRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HandleRestRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Init(t *testing.T) {
	type fields struct {
		Label                       string
		Host                        string
		Port                        uint16
		Token                       string
		conn                        *grpc.ClientConn
		client                      proto.RestInterServiceClient
		registerNewRouteHandler     RegisterNewRouteHandler
		addRouteToIgnoreListHandler AddRouteToIgnoreListHandler
	}
	type args struct {
		config                      *ServiceConfig
		routeRegistrationHandler    RegisterNewRouteHandler
		addRouteToIgnoreListHandler AddRouteToIgnoreListHandler
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Label:                       tt.fields.Label,
				Host:                        tt.fields.Host,
				Port:                        tt.fields.Port,
				Token:                       tt.fields.Token,
				conn:                        tt.fields.conn,
				client:                      tt.fields.client,
				registerNewRouteHandler:     tt.fields.registerNewRouteHandler,
				addRouteToIgnoreListHandler: tt.fields.addRouteToIgnoreListHandler,
			}
			if err := c.Init(tt.args.config, tt.args.routeRegistrationHandler, tt.args.addRouteToIgnoreListHandler); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_ScheduleRestart(t *testing.T) {
	type fields struct {
		Label                       string
		Host                        string
		Port                        uint16
		Token                       string
		conn                        *grpc.ClientConn
		client                      proto.RestInterServiceClient
		registerNewRouteHandler     RegisterNewRouteHandler
		addRouteToIgnoreListHandler AddRouteToIgnoreListHandler
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Label:                       tt.fields.Label,
				Host:                        tt.fields.Host,
				Port:                        tt.fields.Port,
				Token:                       tt.fields.Token,
				conn:                        tt.fields.conn,
				client:                      tt.fields.client,
				registerNewRouteHandler:     tt.fields.registerNewRouteHandler,
				addRouteToIgnoreListHandler: tt.fields.addRouteToIgnoreListHandler,
			}
			c.ScheduleRestart()
		})
	}
}

func TestClient_Start(t *testing.T) {
	type fields struct {
		Label                       string
		Host                        string
		Port                        uint16
		Token                       string
		conn                        *grpc.ClientConn
		client                      proto.RestInterServiceClient
		registerNewRouteHandler     RegisterNewRouteHandler
		addRouteToIgnoreListHandler AddRouteToIgnoreListHandler
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Label:                       tt.fields.Label,
				Host:                        tt.fields.Host,
				Port:                        tt.fields.Port,
				Token:                       tt.fields.Token,
				conn:                        tt.fields.conn,
				client:                      tt.fields.client,
				registerNewRouteHandler:     tt.fields.registerNewRouteHandler,
				addRouteToIgnoreListHandler: tt.fields.addRouteToIgnoreListHandler,
			}
			if err := c.Start(); (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_authenticate(t *testing.T) {
	type fields struct {
		Label                       string
		Host                        string
		Port                        uint16
		Token                       string
		conn                        *grpc.ClientConn
		client                      proto.RestInterServiceClient
		registerNewRouteHandler     RegisterNewRouteHandler
		addRouteToIgnoreListHandler AddRouteToIgnoreListHandler
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Label:                       tt.fields.Label,
				Host:                        tt.fields.Host,
				Port:                        tt.fields.Port,
				Token:                       tt.fields.Token,
				conn:                        tt.fields.conn,
				client:                      tt.fields.client,
				registerNewRouteHandler:     tt.fields.registerNewRouteHandler,
				addRouteToIgnoreListHandler: tt.fields.addRouteToIgnoreListHandler,
			}
			if err := c.authenticate(); (err != nil) != tt.wantErr {
				t.Errorf("authenticate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_requestRestData(t *testing.T) {
	type fields struct {
		Label                       string
		Host                        string
		Port                        uint16
		Token                       string
		conn                        *grpc.ClientConn
		client                      proto.RestInterServiceClient
		registerNewRouteHandler     RegisterNewRouteHandler
		addRouteToIgnoreListHandler AddRouteToIgnoreListHandler
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Label:                       tt.fields.Label,
				Host:                        tt.fields.Host,
				Port:                        tt.fields.Port,
				Token:                       tt.fields.Token,
				conn:                        tt.fields.conn,
				client:                      tt.fields.client,
				registerNewRouteHandler:     tt.fields.registerNewRouteHandler,
				addRouteToIgnoreListHandler: tt.fields.addRouteToIgnoreListHandler,
			}
			if err := c.requestRestData(); (err != nil) != tt.wantErr {
				t.Errorf("requestRestData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
