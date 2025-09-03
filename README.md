# ogbrest

Entry-point for service requests that will be forwarded to appropriate services
within the cluster.

### Connecting other services

Some services may not need to be accessible via REST API, but when they do - they must be configured.
In rest-config.yaml you have to define all the microservices that will be exposed via REST API:
```
services:
- label: label to describe service 
  hostname: localhost
  port: 10001
  token: "random-token-for-the-service"
```

Token should match one defined in the target microservice. Example file contains configuration for every
microservice present in OGB. 