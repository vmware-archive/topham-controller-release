# topham-controller-release

A BOSH release packaging [topham-controller](https://github.com/pivotal-cf-experimental/topham-controller), an OSBAPI-compliant services controller.

## Purpose

A spike exploring a possible implementation of a 'services controller', i.e. a server which understands OSBAPI requests and delegates them to one or more service brokers, while storing the state those brokers emit. Essentially, topham-controller and its CLI replace the service provisioning and management functions in Cloud Controller.

## Usage

Deploy as a BOSH release, ensuring the details of the downstream broker have been set in the manifest ([example](https://github.com/pivotal-cf-experimental/topham-controller-release/blob/master/manifest.yml)).

To connect and provision services, a CLI is provided: see our forked [eden-cli](https://github.com/pivotal-cf-experimental/eden/tree/spike-services-controller-client).
