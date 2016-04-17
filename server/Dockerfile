FROM scratch
COPY server /server
ADD https://curl.haxx.se/ca/cacert.pem /etc/ssl/certs/ca-certificates.crt
EXPOSE 8080
ENTRYPOINT ["/server"]
