HOST ?= 0.0.0.0
API_BASE_URL ?= http://127.0.0.1:8080
LAN_API_BASE_URL ?= http://100.115.97.57:8080
SERVER_BIN ?= server/bin/campus-server
GO_CACHE ?= $(CURDIR)/.cache/go-build

ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: console console1 server server-build test

console:
	cd console && VITE_API_BASE_URL=$(API_BASE_URL) npm run dev -- --host $(HOST)

lan:
	cd console && VITE_API_BASE_URL=$(LAN_API_BASE_URL) npm run dev -- --host $(HOST)

server: server-build
	./$(SERVER_BIN)

server-build:
	mkdir -p $(dir $(SERVER_BIN)) $(GO_CACHE)
	cd server && GOCACHE=$(GO_CACHE) go build -o ../$(SERVER_BIN) .

test:
	mkdir -p $(GO_CACHE)
	cd server && GOCACHE=$(GO_CACHE) go test ./...
	cd console && npm run build
