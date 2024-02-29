PHONY: help

GREEN = \033[32m
RESET = \033[0m

help:
	@echo "🐳 Envoy Sandbox Makefile$(RESET)"
	@echo "-------------------"
	@echo "� $(GREEN)make apps-docker$(RESET) - Build apps docker image"
	@echo "� $(GREEN)make go-proto$(RESET) - Compile proto files"
	@echo "� $(GREEN)make kube-deploy$(RESET) - Deploy the stack to minikube"

check-minikube:
	@if [ -z "$$(minikube status | grep Running)" ]; then \
		echo "❌ $(RED)Minikube is not running$(RESET)"; \
		echo "❌ $(RED)Please run minikube start$(RESET)"; \
		exit 1; \
	fi

kube-deploy: check-minikube
	@kubectl apply -f deploy >/dev/null
	@echo "✅ $(GREEN)envoy stack deployed$(RESET)"

kube-clean: check-minikube
	-@kubectl delete -f deploy 2>/dev/null || true
	@echo "✅ $(GREEN)envoy stack deleted$(RESET)"

apps-docker:
	@echo "🔨 Building apps docker image"
	@docker build -t apps ./apps
	@echo "✅ $(GREEN) Docker image built$(RESET)"

go-proto:
	@echo "🔨 Compiling proto files"
	@protoc -I apps/proto/ apps/proto/hello.proto --go_out=plugins=grpc:apps/proto
	@echo "✅ $(GREEN)Go proto compiled$(RESET)"
