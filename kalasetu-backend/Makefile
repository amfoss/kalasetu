# Run the backend tests against a throwaway Postgres container.
# Or, to use your own server, run:
#   KALASETU_TEST_DATABASE_URL=postgres://user:pw@host:5432/postgres?sslmode=disable go test ./...
PG_CONTAINER := kalasetu-test-pg
PG_PORT ?= 54329

.PHONY: test
test:
	@docker rm -f $(PG_CONTAINER) >/dev/null 2>&1 || true
	@docker run -d --rm --name $(PG_CONTAINER) -e POSTGRES_PASSWORD=test --network host postgres:16-alpine -p $(PG_PORT) >/dev/null
	@until docker exec $(PG_CONTAINER) pg_isready -U postgres -h 127.0.0.1 -p $(PG_PORT) >/dev/null 2>&1; do sleep 0.5; done
	@KALASETU_TEST_DATABASE_URL='postgres://postgres:test@127.0.0.1:$(PG_PORT)/postgres?sslmode=disable' go test -v ./... ; \
		status=$$?; docker rm -f $(PG_CONTAINER) >/dev/null 2>&1; exit $$status
