# Wobscale Payments

## About

This is a stripe api client to let a user subscribe to wobscale services, namely colocation.

## Deploying

Before deploying, make sure your stripe plans exist and are as desired.

Launch a recent CoreOS machine (`899.15.0` at the time of writing).
Make sure the `fleet` and `etcd2` services are enabled and running.

Setup the `config` file referenced in the fleet units.

Run the following:
```
fleetctl load fleets/*.service
fleetctl start payments-ssl.service
# Get a coffee, takes a while for certs to arrive
fleetctl start payments-nginx.service
```

Out of band, you should create the needed DNS records and make sure it all looks okay.
Also, the referenced config files need to exist somehow. Good luck.

## TODO

### Features

* Pay arbitrary $$s for one-time correction, gift, etc.
* Allow users to delete without emailing maybe?
* Loading spinners during ui updates instead of stuff just sorta happening

### UI

* Get rid of all traces of js alerts (ugh).
* Better error handling for server errors.


### Deployment
* Reload certs after an update (sighup nginx)
* route53 presence registration
* not use fleet, k8 or such instead

