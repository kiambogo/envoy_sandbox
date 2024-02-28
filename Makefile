PHONY: help

GREEN = \033[32m
RESET = \033[0m

help:
	@echo "🐳 Envoy Sandbox Makefile$(RESET)"
	@echo "-------------------"
	@echo "🐳 $(GREEN)make docker$(RESET) - Build docker image"
	@echo "🔨 $(GREEN)make go-proto$(RESET) - Compile proto files"

docker:
	@echo "🔨 Building docker image"
	@docker build -t envoy-sandboxy ./apps
	@echo "✅ $(GREEN)Done$(RESET)"

go-proto:
	@echo "🔨 Compiling proto files"
	@protoc -I apps/proto/ apps/proto/hello.proto --go_out=plugins=grpc:apps/proto
	@echo "✅ $(GREEN)Done$(RESET)"
