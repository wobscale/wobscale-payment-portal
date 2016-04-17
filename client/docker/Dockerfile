FROM nginx

COPY docker/entrypoint.sh /entrypoint.sh

RUN mkdir -p /var/www/
WORKDIR /var/www

COPY index.html /var/www/
COPY ./js/  /var/www/js
COPY ./partials/ /var/www/partials
COPY ./css/  /var/www/css
COPY docker/site.conf /etc/nginx/conf.d/

VOLUME /certs
EXPOSE 81
CMD ["sh", "/entrypoint.sh"]
