nuts-registry
#############

Distributed registry for storing and querying health care providers their vendors and technical endpoints.

.. image:: https://circleci.com/gh/nuts-foundation/nuts-registry.svg?style=svg
    :target: https://circleci.com/gh/nuts-foundation/nuts-registry
    :alt: Build Status

.. image:: https://readthedocs.org/projects/nuts-registry/badge/?version=latest
    :target: https://nuts-documentation.readthedocs.io/projects/nuts-registry/en/latest/?badge=latest
    :alt: Documentation Status

.. image:: https://codecov.io/gh/nuts-foundation/nuts-registry/branch/master/graph/badge.svg
    :target: https://codecov.io/gh/nuts-foundation/nuts-registry
    :alt: Code coverage

.. image:: https://api.codeclimate.com/v1/badges/040468237c838c03ff7d/maintainability
   :target: https://codeclimate.com/github/nuts-foundation/nuts-registry/maintainability
   :alt: Maintainability

The registry is written in Go and should be part of nuts-go as an engine.

Dependencies
************

This projects is using go modules, so version > 1.12 is recommended. 1.10 would be a minimum.

Running tests
*************

Tests can be run by executing

.. code-block:: shell

    go test ./...

Building
********

This project is part of https://github.com/nuts-foundation/nuts-go. If you do however would like a binary, just use ``go build``.

The server and client API is generated from the open-api spec:

.. code-block:: shell

    oapi-codegen -generate types,server,client -package api docs/_static/nuts-registry.yaml > api/generated.go

Generating Mocks
****************

These mocks are used by other modules

.. code-block:: shell

    mockgen -destination=mock/mock_client.go -package=mock -source=pkg/registry.go
    mockgen -destination=mock/mock_db.go -package=mock -source=pkg/db/db.go

README
******

The readme is auto-generated from a template and uses the documentation to fill in the blanks.

.. code-block:: shell

    ./generate_readme.sh

This script uses ``rst_include`` which is installed as part of the dependencies for generating the documentation.

Documentation
*************

To generate the documentation, you'll need python3, sphinx and a bunch of other stuff. See :ref:`nuts-documentation-development-documentation`
The documentation can be build by running

.. code-block:: shell

    /docs $ make html

The resulting html will be available from ``docs/_build/html/index.html``

Configuration
*************

Sync modes
==========

The registry supports two modes for updating the internal Db: a file system watcher (``fs``) or downloading from Github (``github``).
When using Github, the registry checks every ``syncInterval`` minutes if anything has changed on Github.
The ``syncAddress`` must point to a tar.gz with the needed registry files included. Github has a nice URL for this.
By default it uses the config in the master branch.

Parameters
==========

The following configuration parameters are available for the registry:

===============================  ===================================================================================  ======================================================================================================================================================
Key                              Default                                                                              Description
===============================  ===================================================================================  ======================================================================================================================================================
clientTimeout                    10                                                                                   Time-out for the client in seconds (e.g. when using the CLI), default: 10
datadir                          ./data                                                                               Location of data files, default: ./data
mode                             server                                                                               server or client, when client it uses the HttpClient, default: server
organisationCertificateValidity  365                                                                                  Number of days organisation certificates are valid, default: 365
syncAddress                      https://codeload.github.com/nuts-foundation/nuts-registry-development/tar.gz/master  The remote url to download the latest registry data from, default: https://codeload.github.com/nuts-foundation/nuts-registry-development/tar.gz/master
syncInterval                     30                                                                                   The interval in minutes between looking for updated registry files on github, default: 30
syncMode                         fs                                                                                   The method for updating the data, 'fs' for a filesystem watch or 'github' for a periodic download, default: fs
vendorCACertificateValidity      1095                                                                                 Number of days vendor CA certificates are valid, default: 1095
===============================  ===================================================================================  ======================================================================================================================================================

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

