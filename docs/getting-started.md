Getting Started
===============

Compile Mothballs
--------------------

Mothballs are compiled, static-content versions of a puzzle category.
You need a mothball for every category you want to run.

To get some mothballs, you'll need to run a development server, which includes the category compiler.
See [development](development.md) for details.


Set up directories
--------------------

    mkdir -p /srv/moth/state
    mkdir -p /srv/moth/mothballs
    cp -r /path/to/src/moth/theme /srv/moth/theme # Skip if using Docker/Podman/Kubernetes

MOTH needs three directories. We recommend putting them all in `/srv/moth`.

* `/srv/moth/state`: (read-write) an empty directory for the server to record its state
* `/srv/moth/mothballs`: (read-only) drop your mothballs here
* `/srv/moth/theme`: (read-only) The HTML5 MOTH client: static content served to web browsers



Run the server
----------------

We're going to assume you put everything in `/srv/moth`, like we suggested.

### Podman

    podman run --name=moth -d -v /srv/moth/mothballs:/mothballs:ro -v /srv/moth/state:/state dirtbags/moth

### Docker

    docker run --name=moth -d -v /srv/moth/mothballs:/mothballs:ro -v /srv/moth/state:/state dirtbags/moth

### Native

    cd /srv/moth
    moth


Copy in some mothballs
-------------------------

    cp category1.mb category2.mb /srv/moth/mothballs

You can add and remove mothballs at any time while the server is running.


Get a list of valid team tokens
-----------------------

    cat /srv/moth/state/tokens.txt

You can edit or replace this file if you want to use different tokens than the pre-generated ones.


Connect to the server
------------------------

Open http://localhost:8080/

Substitute the hostname appropriately if you're a fancypants with a cloud.


Yay!
-------

You should be all set now!

See [administration](administration.md) for how to keep your new MOTH server running the way you want.
