nuts-registry
#############

Distributed registry for storing and querying health care providers their vendors and technical endpoints.

.. image:: https://travis-ci.org/nuts-foundation/nuts-registry.svg?branch=master
    :target: https://travis-ci.org/nuts-foundation/nuts-registry
    :alt: Build Status

.. image:: https://readthedocs.org/projects/nuts-registry/badge/?version=latest
    :target: https://nuts-documentation.readthedocs.io/projects/nuts-registry/en/latest/?badge=latest
    :alt: Documentation Status

.. image:: https://codecov.io/gh/nuts-foundation/nuts-registry/branch/master/graph/badge.svg
    :target: https://codecov.io/gh/nuts-foundation/nuts-registry
    :alt: Code coverage

.. image:: https://api.codacy.com/project/badge/Grade/919adb72a4564722851c7db0ccbec558
    :target: https://www.codacy.com/app/nuts-foundation/nuts-registry
    :alt: Code style

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

    oapi-codegen -package api docs/_static/nuts-registry.yaml > api/generated.go

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

