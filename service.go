package main

import (
	log "github.com/sirupsen/logrus"
)

type Service struct {
	restClients map[string]*Client
}

func (s *Service) Init(r *REST) error {
	s.restClients = make(map[string]*Client)

	log.Infof("Initializing %d REST clients", len(AppConfig.Services))
	for _, service := range AppConfig.Services {
		log.Infof("Initializing REST client %s", service.Label)
		s.restClients[service.Label] = &Client{}
		if err := s.restClients[service.Label].Init(&service, r.RegisterNewRoute); err != nil {
			log.Errorf("Failed to initialize REST client: %s", err.Error())
			return err
		}
	}

	return nil
}

func (s *Service) Start() error {
	// @TODO: Make it safe with reconnects/timeouts and disconnect handling
	for _, client := range s.restClients {
		if err := client.Start(); err != nil {
			log.Errorf("Failed to start REST client: %s", err.Error())
			return err
		}
	}

	return nil
}
