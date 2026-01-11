# Payment Gateway Challenge

This repository contains the Go implementation of the Payment Gateway challenge. If you haven't already read the [README.md](https://github.com/cko-recruitment/) in the root of this organisation, please do so now.


## Running the server

The application is fully containerized. You only need Docker installed (tested with Docker `29.1.1`). To build and start all services:

- `docker compose up`

Once the services are running, the API documentation (Swagger UI) is available at: http://localhost:8090/swagger/index.html

## Running the tests

A CI pipeline runs all tests automatically, but you can also run them locally. Unit tests:

- `go test -v -race ./internal/...`

Integration tests:

- `go test -v ./test/integration/...`

> :warning: ***Important***: To run the integration tests, all services must be running:  `docker compose up -d`


## Design Decisions

### API Endpoints

To meet the functional requirements of the challenge, the API exposes two main endpoints:

- [Create a payment](http://localhost:8090/swagger/index.html#/payments/post_api_v1_payments)
    - This endpoint handles the core functionality of the system, such as authorizing or declining a payment based on the acquiring bank’s response.
    - The rejected scenario was implemented with two assumptions in mind:
        1. If the payment request contains invalid or missing information, the API immediately returns a 400 Bad Request.
        2. If the acquiring bank returns a 4xx error for a valid request, the payment is stored with a rejected status so it can be inspected or handled later if needed.
- [Retrieve payment details by ID](http://localhost:8090/swagger/index.html#/payments/get_api_v1_payments__id_)
    - This endpoint allows merchants to retrieve payment details using an ID. This can be used for reporting purposes or reconciliation processes, especially when a payment was declined or rejected by the acquiring bank.

Both endpoints are implemented synchronously. However, the system could be evolved to make the Create payment flow asynchronous. This would improve throughput and reduce the risk of losing payments under high load. That said, such a change would introduce additional complexity, such as adding a message broker and a mechanism to notify clients about the final payment result.

### Project Structure

I aimed to keep the project structure close to the provided template while making a few pragmatic adjustments:

- Added a `/cmd` directory to organize project binaries. For example, the binary responsible for setting up and running the API server lives here. If a background worker is needed in the future, it could also be added to this directory. This structure follows a well-known Go convention ([reference](https://go.dev/doc/modules/layout#packages-and-commands-in-the-same-repository)).
- Added a `/test/integrations` layer to hold integration tests.
- Kept the `/internal` directory for implementation details, but with some adjustments. The handler package was moved under the api package, since everything inside api is related to the HTTP interface. Handlers act only as entry points and delegate work to the domain layers.
- Introduced `banks` and `payments` domain layers. These layers encapsulate domain logic, interfaces (dependencies), and use cases (services), keeping responsibilities well separated and easier to evolve.

### In-Memory Database

The current implementation uses a simple in-memory data store backed by a `map[PaymentID]Payment`, which allows efficient lookups.

I intentionally followed the suggestion not to include a real database. However, to make the system production-ready, a persistent database would be required. Since the repository relies on abstractions (interfaces) rather than concrete implementations, introducing a real database should be possible with minimal changes to the business logic.

### Developer Experience

To improve the developer experience and reduce friction during development, a few additional improvements were included:

- All components are containerized using Docker, allowing the application and its dependencies to be started with a single command.
- A CI pipeline was added to build and test the application on every pull request or commit to the main branch. This ensures that tests are consistently run and basic issues are caught early.
- A step to automatically generate Swagger documentation was also included, ensuring the API documentation stays up to date and is not forgotten during development.

### Troubleshooting and Best Practices applied

Several best practices were applied to improve reliability, debuggability, and overall robustness of the system:

- A graceful shutdown mechanism was already in place, but a recovery middleware was added to prevent the application from crashing in case of a panic. Instead, the API returns a 500 Internal Server Error, improving system stability.

- Context propagation and request timeouts were implemented across the API. HTTP handlers pass the request context down through the application layers, allowing proper cancellation and timely resource cleanup. This follows a well-known Go pattern used by the standard net/http package and database libraries ([see reference](https://go.dev/blog/context-and-structs)).

- To allow consumers of the payment gateway API to safely retry requests and avoid duplicate charges, a basic idempotency mechanism was implemented on the Create payment endpoint.
Clients can provide an `idempotency_key` in the request payload, and the payment will be processed only once by the acquiring bank.

- The logging setup was improved by adding a middleware that injects a RequestID into each request. This allows logs to be correlated across components, which proved useful during development and would be especially valuable when troubleshooting issues in a production environment.

- A health check endpoint was added to the bank simulator, in addition to the existing ping endpoint. This made it easy to set up a Docker Compose stack to run both projects together and follows a common pattern used in production systems.

- A configuration layer was added to centralize environment variables such as the server port and external dependencies like the Bank Simulator URL.

- The API was versioned so that all endpoints are exposed under /v1. This makes future evolution of the API easier and avoids breaking existing clients.

## What I Intentionally Skipped

The following items were deliberately left out. Some were omitted to keep the scope reasonable for the challenge, while others depend heavily on expected traffic, business priorities, or operational requirements. I’d discuss these with the team if we make this service production-ready:

- Proposing an asynchronous architecture to improve throughput and maximize payment processing rates. While beneficial for the business, this would introduce additional complexity.

- Adding a circuit breaker around the acquiring bank integration. This would improve resilience and protect the system from failures.

- Supporting additional currencies to expand the market reach of the payment gateway.

- Implementing basic observability features such as custom metrics and distributed tracing, which could be achieved using Prometheus and OpenTelemetry.

- Adding basic security. The API currently has no authentication mechanism and does not enforce TLS, which would expose the payment endpoint to man-in-the-middle attacks. These would be mandatory for a production-ready system.

- Introducing a real database to support horizontal scaling and prevent data loss on restarts.

- Adding rate limiting and retry mechanisms for both the payment gateway and the acquiring bank. Network errors are inevitable in production environments and must be handled gracefully.

- Introducing a caching layer for frequently accessed read operations, depending on the expected read patterns and workload from merchants.
