.. _nuts-registry-configuration:

Nuts registry configuration
###########################

.. marker-for-readme

Sync modes
==========

The registry supports two modes for updating the internal Db: a file system watcher (``fs``) or downloading from Github (``github``).
When using Github, the registry checks every ``syncInterval`` minutes if anything has changed on Github.
The ``syncAddress`` must point to a tar.gz with the needed registry files included. Github has a nice URL for this.
By default it uses the config in the master branch.

The following configuration parameters are available for the registry.

===================================     ====================================================================================================    ================================================================================
Key                                     Default                 Description
===================================     ====================================================================================================    ================================================================================
registry.datadir                        ./data                                                                                                  Location of data files
registry.mode                           server                                                                                                  server or client, when client it uses the HttpClient
registry.address                        localhost:1323                                                                                          Interface and port for http server to bind to
registry.syncMode                       fs                                                                                                      ``fs`` or ``github``
registry.syncInterval                   30                                                                                                      Interval in minutes to check for new registry data on github
registry.syncAddress                    https://codeload.github.com/nuts-foundation/nuts-registry-development/tar.gz/master                     The tar.gz to download from github
===================================     ====================================================================================================    ================================================================================

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

