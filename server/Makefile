# TODO, this doesn't need to be phony, we can be faster and only build when changed
# That'll help with docker's caching a lot too and speed it up even more!
.PHONY: server
server:
	docker run --rm -v "$(shell pwd):/go/src/github.com/wobscale/wobscale-payment-portal/server" --workdir=/go/src/github.com/wobscale/wobscale-payment-portal/server golang:1.10 make static

.PHONY: static
static:
	CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' -o server server.go

.PHONY: server-docker
server-docker: server ./files/cacert.pem
	docker build -t euank/wobscale-payments-server -f ./server/Dockerfile ./server

# cacert.pem is checked in for our static dockerfile because downloading of
# them on e.g. travis breaks due to curl.haxx.se doing heavy rate-limiting of
# the file.
# It's nicer to them to just 
./files/cacert.pem:
	wget -O files/cacert.pem https://curl.haxx.se/ca/cacert.pem
