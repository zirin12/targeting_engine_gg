# targeting_engine_gg
Targeting Engine that will route the right campaigns to the right requests.

---

## How to Run

Ensure MongoDB is running, then:

```bash
go run cmd/server/main.go
```
Default settings:
- Port: 8080
- Mongo URI: mongodb://localhost:27017

## API Endpoint
### GET /v1/delivery
Returns campaigns that match provided dimensions:

```curl 'http://localhost:8080/v1/delivery?country=IN&app=com.foo.app&os=android'```

### Response:
- Query keys are validated against ValidDimensions()
- Missing required keys return 400 Bad Request
- If no match is found: 204 No Content
- Successful match: JSON array of campaigns

## Future Improvements
- Structured logging and tracing
- Unit tests for matcher and store logic
- Prometheus metrics (QPS, match rate)
- API Middleware
- Dependency injection wherever possible
- YAML/JSON config loader for tenant-based dimensions
- Router upgrade (chi or mux) for flexible routing
- Use map for storing dimensions that can simplify rules check and also fast lookups
- Use Redis as the in memory store instead of a slice of campaigns
- Maybe have the entire slice as an atomic unit that will be swapped for each update by sync/poller, avoiding all the read/write locks
- If updates are frequent poller/sync job could push it a queue from where it can be consumed so poller is free

## Legacy Prototype
The original main.go is preserved in /legacy/main.go

Useful for rapid prototyping or onboarding. Can be run via:
```go run prototype/main.go```




