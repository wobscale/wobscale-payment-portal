#!/bin/bash

for var in ENV_GITHUB_CLIENT_ID ENV_ENVIRONMENT ENV_STRIPE_PUBLISHABLE_KEY; do
  # indirect reference; get the real environment variable
  # http://tldp.org/LDP/abs/html/ivr.html
  eval v=\$$var
  sed -i "s/$var/$v/g" index.html
done

exec nginx -g 'daemon off;'
