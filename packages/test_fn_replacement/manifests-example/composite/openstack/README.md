Composite: openstack
====================

This composite assembles openstack functions into a functional cloud platform.

By default the functions will use default (likely incorrect) endpoint
information.  To inject appropriate endpoint information, the following should
be done in lower kustomizations:

* pull in this composite, which defines the `openstack-endpoint-catalogue`
* patch or replace any desired external endpoints (e.g. fluentd) into
  `openstack-endpoint-catalogue`
* (TODO) inject TLS information into `openstack-endpoint-catalogue`
* supply a `site-networking-catalog` that contains a spec.DOMAIN
* apply the `replacements-endpoints` kustomization in this composite to
  construct the final endpoints and inject them into the functions
