.. _nuts-registry-configuration:

Nuts registry configuration
###########################

.. marker-for-readme

The following configuration parameters are available for the registry.

===================================     ====================    ================================================================================
Key                                     Default                 Description
===================================     ====================    ================================================================================
registry.datadir                        ./data                  Location of data files
registry.mode                           server                  server or client, when client it uses the HttpClient
registry.address                        localhost:1323          Interface and port for http server to bind to
===================================     ====================    ================================================================================

As with all other properties for nuts-go, they can be set through yaml:

.. sourcecode:: yaml

    registry:
       datadir: ./data

as commandline property

.. sourcecode:: shell

    ./nuts --registry.datadir ./data

Or by using environment variables

.. sourcecode:: shell

    NUTS_REGISTRY_DATADIR=./data ./nuts

