.PHONY: all
all: ;

.PHONY: pg
pg:
	docker run --rm \
		--name=go-metrics-db \
		-v $(abspath ./db/init/):/docker-entrypoint-initdb.d \
		-v $(abspath ./db/data/):/var/lib/postgresql/data \
		-e POSTGRES_PASSWORD="P@ssw0rd" \
		-d \
		-p 5432:5432 \
		postgres:16.3

.PHONY: stop-pg
stop-pg:
	docker stop go-metrics-db

.PHONY: clean-data
clean-data:
	sudo rm -rf ./db/data/

.PHONY: lint
lint:
	golangci-lint run --fix

.PHONY: pprofdiff
pprofdiff:
	go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof

.PHONY: build-agent
build-agent:
	go build -o build/agent -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$$(date +%Y-%m-%d) -X main.buildCommit=$$(git rev-parse HEAD)" ./cmd/agent

.PHONY: build-server
build-server:
	go build -o build/server -ldflags "-X main.buildVersion=v1.0.0 -X main.buildDate=$$(date +%Y-%m-%d) -X main.buildCommit=$$(git rev-parse HEAD)" ./cmd/server

.PHONY: test
test:
	go test -v ./...
