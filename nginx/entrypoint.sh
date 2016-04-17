#!/bin/bash

for var in ENV_API_PORT ENV_API_NAME ENV_WEB_PORT ENV_WEB_NAME; do
  # indirect reference; get the real environment variable
  # http://tldp.org/LDP/abs/html/ivr.html
  eval v=\$$var
  sed -i "s/$var/$v/g" /etc/nginx/conf.d/site.conf
done

exec nginx -g 'daemon off;'
