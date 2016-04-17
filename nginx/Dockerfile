FROM nginx

COPY entrypoint.sh /entrypoint.sh

COPY site.conf /etc/nginx/conf.d/
VOLUME /certs

CMD ["sh", "/entrypoint.sh"]
