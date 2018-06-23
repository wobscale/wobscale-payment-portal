FROM golang:1.10 as builder
COPY ./ /go/src/github.com/wobscale/wobscale-payment-portal/
WORKDIR /go/src/github.com/wobscale/wobscale-payment-portal/server/
RUN make static

FROM scratch
COPY --from=builder /go/src/github.com/wobscale/wobscale-payment-portal/server/server /server
COPY ./server/files/cacert.pem /etc/ssl/certs/ca-certificates.crt
EXPOSE 8080
ENTRYPOINT ["/server"]
