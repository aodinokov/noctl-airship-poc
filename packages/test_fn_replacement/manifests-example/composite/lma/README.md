Composite: lma
==============

This composite assembles logging, monitoring, and alerting functions that
will generally be desired in non-development use cases.

By default the functions will use default (likely incorrect) endpoint
information.  To inject appropriate endpoint information, the following should
be done in lower kustomizations:

* pull in this composite, which defines the `lma-endpoint-catalogue`
* patch or replace any desired external endpoints into
  `lma-endpoint-catalogue`
* (TODO) inject TLS information into `lma-endpoint-catalogue`
* supply a `site-networking-catalog` that contains a spec.DOMAIN
* apply the `replacements-endpoints` kustomization in this composite to
  construct the final endpoints and inject them into the functions
