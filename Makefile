.PHONY: build \
		clean \
		postgres-up postgres-start postgres-stop postgres-rm postgres-connect \
		run

POSTGRES_USER=gophermart
POSTGRES_PASSWORD=secret
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=gophermartdb
POSTGRES_DSN=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

APP=cmd/gophermart/gophermart

build:
	go build -o $(APP) ./cmd/gophermart

clean:
	rm -f $(APP)

run: build
	./$(APP)

postgres-up:
	docker run --name gophermart-postgres \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		-p $(POSTGRES_PORT):5432 \
		-d postgres:16

postgres-start:
	docker start gophermart-postgres

postgres-stop:
	docker stop gophermart-postgres

postgres-rm:
	docker rm -f gophermart-postgres

postgres-connect:
	docker exec -it gophermart-postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)
