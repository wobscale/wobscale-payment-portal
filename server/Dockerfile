FROM scratch
COPY server /server
COPY ./files/cacert.pem /etc/ssl/certs/ca-certificates.crt
EXPOSE 8080
ENTRYPOINT ["/server"]
