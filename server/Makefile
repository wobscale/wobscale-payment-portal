.PHONY: static
static:
	CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' -o server server.go

DOCKER_PREFIX := wobscale/payments
DOCKER_TAG := $(shell git rev-parse --short HEAD)

.PHONY: server-docker
server-docker: ./files/cacert.pem
	docker build -f ./Dockerfile -t $(DOCKER_PREFIX)-server:$(DOCKER_TAG) ../

# cacert.pem is checked in for our static dockerfile because downloading of
# them on e.g. travis breaks due to curl.haxx.se doing heavy rate-limiting of
# the file.
# It's nicer to them to just 
./files/cacert.pem:
	wget -O files/cacert.pem https://curl.haxx.se/ca/cacert.pem

clean:
	rm -f ./files/cacert.pem server
