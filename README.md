# Wobscale Payments

## About

This is a stripe api client to let a user subscribe to wobscale services, namely colocation.

## Deploying

Before deploying, make sure your stripe plans exist and are as desired.

Then create a Kubernetes (some assembly required).

Finally, use kubectl to create resources similar to [these](https://github.com/euank/ek8s/tree/master/wobscale-paypi).

## TODO

### Features

* Pay arbitrary $$s for one-time correction, gift, etc.
* Allow users to delete without emailing maybe?
* Loading spinners during ui updates instead of stuff just sorta happening

### UI

* Get rid of all traces of js alerts (ugh).
* Better error handling for server errors.


### Deployment
* route53 presence registration
