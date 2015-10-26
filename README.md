vindaloo [![Build Status](https://travis-ci.org/euforia/vindaloo.svg?branch=master)](https://travis-ci.org/euforia/vindaloo)
========
An inventory system to store, query and analyze infrastructure, configurations or any other type of data.

The system can be useful to store/query/analyze information such as configurations for loadbalancers, containers, virtual machines, physical hardware, cloud infrastructure/components or any other kind of data you see fit.  Each asset has an associated type and can contain any kind of arbitrary given the system is schema less.


# Table of Contents
1. [Installation](#installation)
2. [Configuration](#configuration)
3. [Auth Tokens](#auth-tokens)
4. [User Guide](#user-guide)
5. [Local Auth Groups](#local-auth-groups)


Installation
------------
System packages are available for installation.  You can also manually build the project.

#### Requirements

    go >= 1.4.2 (not needed for packaged install)
    elasticsearch >= 1.4.x

##### Package install
`rpm` and `deb` packages can be found on [packagecloud.io](https://packagecloud.io/euforia/vindaloo).  Installation instructions are also available there if needed.

##### Manual install
To perform a manual install continue to the section below:

    # Build and install
    $ go get github.com/euforia/vindaloo
    $ cd $GOPATH/src/github.com/euforia/vindaloo
    $ make all
    
This will install the the directory structure under `./build/opt/vindaloo/`, which should be copied to `/opt/vindaloo`.  Once copied change to that directory and proceed with configuration section.

Configuration
-------------
Start by copying the sample configuration files in the `etc` directory.

    $ cp etc/vindaloo.json.sample etc/vindaloo.json
    $ cp etc/local-groups.json.sample etc/local-groups.json

Fill in the appropriate values keeping in mind that paths in the config file that do not start with `/` are treated as relative paths from the current working directory. After your configuration is complete you can start the service.  Here's a description of the various configuration options.

##### auth
Auth is required for all write requests as well as to listen to event subscriptions.  By default `basic` http auth is enabled using the `etc/htpasswd` file.  The default credentials are `admin:vindaloo`.  

The `groups_file` is used to specify which users are allowed to create asset types as per the `LocalAuthGroups`.  A sample config file is located under `etc/`.

LDAP auth is also available. This can be used by setting the `type` field to `ldap` and adding the appropriate options in the `config` section.  The available options are `url`, `search_base`, `cache_ttl`, `bind_dn`, and `bind_password`.  `bind_password` can either be a string or a location to a file containing the password.  In the case of a file the path must be proceeded by `file://` e.g (default). 

    file:///opt/vindaloo/etc/bindpasswd

##### datastore
Currently, elasticsearch is the only supported backend.  The only values that may require modifying are `host` and `port` based on your setup.

##### endpoints
Endpoint configurations.

##### asset
The `required_fields` specifies the fields that are required for any given asset (default: status, environment). Although more field requirments can be added, the default keys should not be removed.

The `enforced_fields` specifies fields that can only contain the specified values (default: status, environment).

##### default\_result\_size
This is the number of results that will be returned when the `size` parameter is not specified. (default: 100)

##### webroot
Path to the web directory.

The `version` field should not be touched. Changing this will prevent the process from starting up.

Starting the Service
--------------------
Once the installation and configuration is complete, you'll need to make sure elasticsearch is restarted.  To do so run the following as necessary:

    $ /etc/init.d/elasticsearch restart

Once elasticsearch is running (it can take some time to become available), execute the following command to start the service:

    $ ./bin/vindaloo-ctl start

You can now start using the system.


User Guide
----------
Here you can find the basics and fundamentals to start using the inventory system.

### Asset Type
Every asset has a asset type.  To avoid the creation of unwanted types, only users in the `admin` group of the `LocalAuthGroups` are allowed to create asset types.  The type will automatically be created upon the creation of the asset if done by an authorized `admin` user.  More information about the `LocalAuthGroups` can be found below.


### Versions
Versions are automatically created on each write.  When a write occurs, the existing asset is copied over to the `versions` index incrementing the version number then performing the write.  It is possible get a list of versions or specific versions of a given asset.  A version to version diff can also be obtained.


### Asset
Each asset must have an associated type.  Before an asset can be created, the asset type must be created.  As mentioned before, only admins are allowed to create new asset types. An asset has versions available.  These are only available after the first write operation.  A `current` asset has no version.  

### Events
If enabled events are fired on all `write` actions.  The available event types are:

Event type         | Payload 
-------------------|---------------------
assettype.created  | Asset type id
asset.created      | Complete asset data
asset.updated      | Updated asset data
asset.deleted      | Asset id

Below are samples for each type:

##### assettype.created

    {
        "type": "assettype.created",
        "timestamp": "....",
        "payload": {
            "id": "loadbalancer"
        }
    } 

##### asset.created

    {
        "type": "asset.created",
        "timestamp": "....",
        "payload": {
            "id": "foo.bar.org",
            "type": "virtualserver",
            "timestamp": "...",
            "data": {
                "user_supplied": "data_on_creation"
            }
        }
    }

##### asset.updated

    {
        "type": "asset.updated",
        "timestamp": "....",
        "payload": {
            "id": "foo.bar.org",
            "type": "virtualserver",
            "timestamp": "...",
            "data": {
                "data_that": "was_updated"
            }
        }
    }

##### asset.deleted

    {
        "type": "asset.deleted",
        "timestamp": "....",
        "payload": {
            "id": "foo.bar.org"
        }
    }   


### Endpoints
The following verbs and endpoints are available:

| Endpoint                                  | Method  | Description
| ----------------------------------------- | ------- | -----------
| **/v3/**                                  | GET     | List types
|                                           | OPTIONS | Get ACL's and usage
| **/v3/< asset_type >**                    | GET     | List/Search within a type
|                                           | POST    | Create type
|                                           | OPTIONS | Get ACL's and usage
| **/v3/< asset_type >/properties**         | GET     | Get properties for type
| **/v3/< asset_type >/< asset >**          | GET     | Get 
|                                           | POST    | Create
|                                           | PUT     | Update
|                                           | DELETE  | Remove
|                                           | OPTIONS | Get ACL's and usage
| **/v3/< asset_type >/< asset >/versions** | GET     | Get versions
|                                           | OPTIONS | Get ACL's and usage
| **/v3/raw**                               | GET     | Pass-through request to elasticsearch index
| **/v3/raw/versions**                      | GET     | Pass-through request to elasticsearch versions index
| **/v3/events/< event >**                  | N/A     | Websocket to subscribe to events.
| **/v3/search**                            | GET     | Search
| **/config**                               | GET     | Get config
| **/auth/access_token**                    | POST    | Get access token

##### List types

    - GET /v3/

Response e.g.:
    
    [
        {"name": "virtualserver", "count": 1233}, 
        {"name": "dnsrecord", "count": 1543}
    ]

##### List properties for type

    - GET /v3/<asset_type>/properties

Response e.g.:
    
    [
       "id",
       "timestamp",
       "created_by",
       "updated_by",
       "environment",
       "status",
       ...
    ]

##### Get asset

    - GET /v3/<asset_type>/<asset_id>

Response e.g.:

    {
        "id": "foo.bar.org"
        "type":"virtualserver",
        "timestamp": <epoch>,
        "data":{
            "status":"running",
            "environment": "dev",
            "created_by": "user1",
            "updated_by": "user2"
            ....
        }
    }

##### Get asset version

    - GET /v3/<asset_type>/<asset_id>?version=<version>

The response is the same as an asset but also includes the version attribute.

Response e.g.:

    {
        "id": "<asset_id>"
        "type":"<asset_type>",
        "timestamp": <epoch>,
        "data":{
            "status":"running",
            "environment": "dev",
            "version": <version>,
            "created_by": "user1",
            "updated_by": "user2"
            ....
        }
    }

##### Get asset versions
Asset versions can be obtained by calling the following endpoint.

    - GET /v3/<asset_type>/<asset_id>/versions

Additionally version to version incremental diffs can also be obtained using the `diff` parameter.

    - GET /v3/<asset_type>/<asset_id>/versions?diff

Response e.g.:

    [{
        "version": 2,
        "against_version": 1,
        "updated_by": "....."
        "timestamp": <time_value>
        diff: "<diff_data>"
    },{
        ....
    }]

##### Create new asset
When creating an asset 2 fields are required - `status` and `environment` or as specified in your config.  When creating an asset 2 additional fields are automatically added - `created_by` and `updated_by` with the user specified as part of the auth.

    - POST /v3/<asset_type>/<asset_id>

        {
            "name": "foo.bar.com",
            "status": "running",
            "environment": "development"
            ...
        }

Response e.g.:

    { "id": "<asset_id>" }
    
##### Edit existing asset
Editing an asset will also update the `updated_by` field with the authenticated user.

    - PUT /v3/<asset_type>/<asset_id>?delete_fields=foo,bar

        {
            "status": "stopped"
            "description": null
            ...
        }

Response e.g.:

    { "id": "<asset_id>" }

In the above example we update the `status` field and delete the fields called `foo` and `bar`.

##### Delete asset

    - DELETE /v3/<asset_type>/<asset_id>

##### Search for asset

As a request body:

    - GET /v3/<asset_type>

        {
            "status": "stopped",
            "os": "ubuntu"
        }

As query parameters:

    - GET /v3/<asset_type>?status=stopped&os=ubuntu

This matches both attributes.  Additionally the following parameters are also available:

* **sort**: Sort the result by the given attribute in ascending or descending order (e.g. sort=name:asc *or* sort=name:desc)
    
* **from**: This can be used for pagination to specify the offset (e.g. from=0)
    
* **size**: Number of results to return from the offset `from` if specified (e.g. size=100)

* **aggregate**: This is used to aggregate counts of a given field.  For instance, for a field called `os` with values `centos` and `ubuntu`, to get a distinct count of values you would set the aggregator to `os`.

For example:

    GET /v3/<asset_type>?aggregate=os

Response:

    [{
        "name": "centos",
        "count": 200
    },{
        "name": "ubuntu",
        "count": 123
    }]

##### Subscribe to events
In order to subscribe to events, an access token must be generated.  See below for more details on how to generate tokens.  

Once a token has been generated, you can subscribe to events by connecting to the websocket endpoint - `ws://localhost:5454/v3/events/< event >`, where `event` can be a an event type or an event type wild card.

Subscribe to a specific event (e.g):

    ws://localhost:5454/v3/events/asset.created

Subscribe to all `asset` events (e.g):

    ws://localhost:5454/v3/events/asset.*

Subscribe to all events (e.g):

    ws://localhost:5454/v3/events/.*

Example (javascript):

    // Subscribe to all asset related events
    var ws = new WebSocket('ws://localhost:5454/v3/events/asset.*?access_token=<my_token>');

    ws.on('open', function() { console.log('opened') });

    // Set callback for event.
    ws.on('message', function(message) {
        var data = JSON.parse(message);
        console.log('Received event:', data);
        // DO SOMETHING WITH THE DATA
    });

    ws.on('close', function() { console.log('closed'); });

Auth Tokens
-----------
To use token authentication, you need generate a token with HTTP basic authentication (default login is admin/vindaloo):

    - POST /auth/access_token

Example:

    $ curl -XPOST localhost:5454/auth/access_token -u admin
    Enter host password for user 'admin':

Response:

    { "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiZXhwIjoxNDY3Njc2MDc2LCJpYXQiOjE0NDE3NTYwNzYsImlzcyI6InZpbmRhbG9vIiwic3ViIjoiYWRtaW4ifQ.gYN5tSk-ZYKQ--2w4NzuJ1Zt-VAnzOSCdDbl3JJK7P3" }

You can now use this token as a url parameter or in the header.  

##### URL Parameter
To use it as a url parameter, you need to use the `access_token`.  Here is an example to create a new asset using the `access_token` parameter:
    
    curl -XPOST localhost:5454/v3/serviceprofile/my_profile?access_token=<token> -d '{
        "status": "enabled",
        "environment": "production"
    }'

##### Request Header
The same request as above using headers, would look like this:

    curl -XPOST -H "Authorization: BEARER <token>" localhost:5454/v3/serviceprofile/my_profile -d '{
        "status":" enabled",
        "environment": "production"
    }'


Local Auth Groups
-----------------
Local auth groups are primarily used to create asset types.  The configuration file can be found at etc/local-groups.json.  Even though auth is disabled, a simple lookup against the `local-groups.json` file to prevent accidental asset type creation.  

Add the usernames to the `admin` field you wish to allow.  The user must match that used for 'HTTP Basic Auth'.  The password can be left blank.  If auth is enabled, then actual authentication with the backend will be performed.

Example:

    {
        "admin": [ "user1", "user2" ]
    }

Development
-----------
In order to setup the development environment you'll need the following:

    - golang >= 1.4.2
    - make
    - docker

To perform a local (native) build you can simply run:

    make all

Assuming all requirements are met, a full linux build can be performed as follows:

    ./scripts/build.sh

This will produce the following:

- Linux binaries and other necessary files under the `build` folder

- A .rpm and .deb in the `build` folder

- A docker image called `euforia/vindaloo`

###### Notes:

- Testing has primarily been done on Oracle 6.6/7 though it should work on any OS given the requirements are met.
