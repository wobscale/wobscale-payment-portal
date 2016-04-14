FROM nginx

COPY docker/entrypoint.sh /entrypoint.sh

RUN mkdir -p /var/www/
WORKDIR /var/www

COPY css js index.html css partials /var/www/
COPY docker/site.conf /etc/nginx/conf.d/

VOLUME /certs
CMD ["sh", "/entrypoint.sh"]
