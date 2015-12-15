0.5.2
-----
* Refactored abstraction layers

0.5.1
-----
* Lowered generic string type mapping precedence for default mapping.
* Fix for aggregate counts
* Added Godep support

0.5.0
-----
* Project moved to `vindalu/vindalu` on github.
* Added constraints on resource type and id values.
* Code refactor.

0.4.5
-----
* Support for multiple mappings based on type.

0.4.4
-----
* Added search handler.
* Fixed raw handler to only allow access to `vindalu` indexes.
* Updated default elasticsearch mappings.
* Removed AWS plugin.
* Fix for publish call to return error.
* Added `/status` endpoint.

0.4.3
-----
* Fix for deleting fields without a POST body.

0.4.2
-----
* Changed field deletion to be a request parameter `delete_fields`.
* Added base plugin framework.
* Added `aws` plugin to populate ec2 instance/s.

0.4.1
-----
* UI moved to a separate project.
* Added `created_on` field to track creation date and time.
* Fix for incorrect `timestamp` when creating a new version.

0.4.0
-----
* Added token auth support with `signing_key` and `ttl` configurables.
* Added http basic auth support using `htpasswd` files.
* Refactor of auth flow to support http basic auth.
* Added token auth to event subscriptions.
* Added edit and delete functionality to UI with auth.
* Fix for version skipping/jumping.
* Fix for token ttl being too long.

0.3.1
-----
* Added event subscription via websocket endpoint `/events/< topic >`.
* Added CI build to push packages from master.
* Added package deployments to packagecloud.io

0.3.0
-----
* Create asset types by calling `/< asset_type >` with optional `properties`
* Added `nats` as the messaging and queueing backend for write requests.
* Added `/< asset_type >/properties` endpoint to list all properties for a type.
* Added OPTIONS call to return basic endpoint usage.
