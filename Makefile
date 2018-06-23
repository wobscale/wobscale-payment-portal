all: server-docker client-docker

# docker prefix, e.g. quay.io/username/repo
# The suffixes '-server:$TAG' and '-client:$TAG' will be appended to it and used as repos
DOCKER_PREFIX := wobscale/payments

DOCKER_TAG := $(shell git rev-parse --short HEAD)

.PHONY: server-docker
server-docker:
	make -C server server-docker DOCKER_PREFIX=$(DOCKER_PREFIX) DOCKER_TAG=$(DOCKER_TAG)

# Note: this docker image is only used for local development, not production
# The tag is left out for that reason
.PHONY: nginx
nginx:
	docker build -t $(DOCKER_PREFIX)-nginx -f ./nginx/Dockerfile ./nginx

.PHONY: client-docker
client-docker:
	docker build -t $(DOCKER_PREFIX)-client:$(DOCKER_TAG) -f ./client/docker/Dockerfile ./client

# NOTE, this port must match your github clientid/secret's 'redirect_uri'; the
# 'redirect uri' should be https://127.0.0.1:$DEVPORT/login
DEVPORT := 12443
# Override with make DEVPORT=$port

.PHONY: dev-checkenv
dev-checkenv:
ifndef STRIPE_API_KEY
	$(error STRIPE_API_KEY required)
endif
ifndef GITHUB_SECRET_KEY
	$(error GITHUB_SECRET_KEY required)
endif
ifndef GITHUB_CLIENT_ID
	$(error GITHUB_CLIENT_ID required)
endif
ifndef STRIPE_PUBLISHABLE_KEY
	$(error STRIPE_PUBLISHABLE_KEY required)
endif

# Dev use only
./certs/ssl.key ./certs/ssl.pem:
	mkdir -p certs
	openssl req -subj '/CN=127.0.0.1:$(DEVPORT)/O=wobscale/C=US/subjectAltName=127.0.0.2:$(DEVPORT)' \
	  -new -newkey rsa:2048 -sha256 -days 365 -nodes -x509 \
	  -keyout certs/ssl.key -out certs/ssl.pem

./certs/dhparam.pem:
	mkdir -p certs
	openssl dhparam -out ./certs/dhparam.pem 2048

.PHONY: certs
certs: ./certs/dhparam.pem ./certs/ssl.key ./certs/ssl.pem

.PHONY: dev
dev: dev-checkenv nginx client-docker server-docker certs
	-docker network create wobscale-payments
	-docker rm --force "wobscale-payments-server"
	-docker rm --force "wobscale-payments-client"
	-docker rm --force "wobscale-payments-nginx"
	docker run -d --net=wobscale-payments --name="wobscale-payments-server" \
	           -e ENV_ENVIRONMENT=dev \
	           -e STRIPE_API_KEY=$(STRIPE_API_KEY) \
	           -e GITHUB_SECRET_KEY=$(GITHUB_SECRET_KEY) \
	           -e GITHUB_CLIENT_ID=$(GITHUB_CLIENT_ID) \
	           -e CORS_ALLOW_ORIGIN="*" \
	           $(DOCKER_PREFIX)-server:$(DOCKER_TAG)
	docker run -d --net=wobscale-payments --name="wobscale-payments-client" \
	           -e ENV_ENVIRONMENT=dev \
	           -e ENV_STRIPE_PUBLISHABLE_KEY=$(STRIPE_PUBLISHABLE_KEY) \
	           -e ENV_GITHUB_CLIENT_ID=$(GITHUB_CLIENT_ID) \
	           -e ENV_API_URL=https://127.0.0.2:$(DEVPORT) \
	           $(DOCKER_PREFIX)-client:$(DOCKER_TAG)
	sleep 3
	docker run -d --net=wobscale-payments --name="wobscale-payments-nginx" -p $(DEVPORT):443 \
	           -e ENV_API_PORT=443 \
	           -e ENV_WEB_PORT=443 \
	           -e ENV_API_NAME=127.0.0.2 \
	           -e ENV_WEB_NAME=127.0.0.1 \
	           -e ENV_CLIENT_NAME=wobscale-payments-client \
	           -e ENV_SERVER_NAME=wobscale-payments-server \
	           -v "$(shell pwd)/certs:/certs" \
	           $(DOCKER_PREFIX)-nginx
	@echo "Visit https://127.0.0.2:$(DEVPORT) and accept an ssl warning.."
	@echo "Then visit https://127.0.0.1:$(DEVPORT) for a good time :)"

push:
	docker push $(DOCKER_PREFIX)-client:$(DOCKER_TAG)
	docker push $(DOCKER_PREFIX)-server:$(DOCKER_TAG)
