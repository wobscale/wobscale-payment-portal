# TODO, this doesn't need to be phony, we can be faster and only build when changed
# That'll help with docker's caching a lot too and speed it up even more!
.PHONY: server
server:
	docker run --rm -v "$(shell pwd):/go/src/github.com/wobscale/wobscale-payment-portal/server" --workdir=/go/src/github.com/wobscale/wobscale-payment-portal/server golang:1.7 make static

.PHONY: static
static:
	CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' -o server server.go

.PHONY: server-docker
server-docker: server
	docker build -t euank/wobscale-payments-server -f ./server/Dockerfile ./server
