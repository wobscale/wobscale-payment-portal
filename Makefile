all: server-docker client-docker

.PHONY: server
server:
	CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' -o ./server/server ./server/server.go

.PHONY: server-docker
server-docker: server
	docker build -t euank/wobscale-payments-server -f ./server/Dockerfile ./server
	docker build -t euank/wobscale-payments-server-nginx -f ./server/nginx/Dockerfile ./server/nginx

.PHONY: client-docker
client-docker:
	docker build -t euank/wobscale-payments-client -f ./client/docker/Dockerfile ./client
