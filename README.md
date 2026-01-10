# Instructions for candidates

This is the Go version of the Payment Gateway challenge. If you haven't already read the [README.md](https://github.com/cko-recruitment/) in the root of this organisation, please do so now.

## Implementation



Example 

```
curl -v -X 'POST' \                          gorila-eks-prod
  'http://localhost:8090/api/v1/payments' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "amount": 1000,
  "card_number": "4111111111111111",
  "currency": "USD",
  "cvv": "123",
  "expiry_month": 12,
  "expiry_year": 2050
}'

{"time":"2026-01-10T17:49:09.561688974Z","level":"WARN","msg":"payment status is not authorized","method":"POST","path":"/api/v1/payments","request_id":"1a767247e2ba/Hb8Ia1FcLJ-000001","payment_id":"019ba906-bc39-7a6c-97da-e66750abdc28","payment_status":"rejected"}

```

## Generate Doc

- versioning [x]
- including context [x]
- timeout request [x]
- logging with context [x]
- rejected [x]
- currency business rules [x] contract enforcement!!!
- idempotency key [x] ???

- ci/cd [] with build and tests containerized

- e2e []

- include logging requestID middleware, that injects it an ID to facilitate troubleshooting [troubleshooting, production readiness]
- include recoverer middleware, that recover the api in case of panics returning internal error instead of crashing [troubleshooting, production readiness]
- include timeout middleware, that will returns 504 gateway timeout, defualts to 60 seconds, so it will not using resources undefinetly [troubleshooting, production readiness]
- middleware logging [troubleshooting, production readinesss, dx]
- docker compose -- dev experience, bank simulator healthcheck, api containarized, auto generate documentation (docker compose run --rm swagger-gen-docs) [dx]

- docker run --rm -v $(pwd):/code ghcr.io/swaggo/swag:latest init -g cmd/api/main.go

- unit test
- integration test
- e2e test
- versioning
- monitoring
- security
- Interfaces to test payment gateway and database
- Idempotency ? (payment gateway)
- Rate Limiter ? (bank)
- Circuit Breaker ? (bank)

Notes for my future self:

- I tried to apply some knowledge about DDD but what about banks? the interface is living in the implementation not in the consumer, repository I did the opposite...


## Template structure
```
main.go - a skeleton Payment Gateway API
imposters/ - contains the bank simulator configuration. Don't change this
docs/docs.go - Generated file by Swaggo
.editorconfig - don't change this. It ensures a consistent set of rules for submissions when reformatting code
docker-compose.yml - configures the bank simulator
.goreleaser.yml - Goreleaser configuration
```

Feel free to change the structure of the solution, use a different test library etc.

### Swagger
This template uses Swaggo to autodocument the API and create a Swagger spec. The Swagger UI is available at http://localhost:8090/swagger/index.html.
