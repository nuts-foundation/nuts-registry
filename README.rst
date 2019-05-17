nuts-registry
=============

Distributed registry for storing and querying healthcare care providers their vendors and technical endpoints.

.. image:: https://travis-ci.org/nuts-foundation/nuts-registry.svg?branch=master
    :target: https://travis-ci.org/nuts-foundation/nuts-registry
    :alt: Build Status

.. image:: https://readthedocs.org/projects/nuts-registry/badge/?version=latest
    :target: https://nuts-documentation.readthedocs.io/projects/nuts-registry/en/latest/?badge=latest
    :alt: Documentation Status

.. image:: https://codecov.io/gh/nuts-foundation/nuts-registry/branch/master/graph/badge.svg
    :target: https://codecov.io/gh/nuts-foundation/nuts-registry

.. inclusion-marker-for-contribution

To generate the Server stub install some dependencies:

.. code-block:: shell

   go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen

Then run

.. code-block:: shell

   oapi-codegen -generate server PATH_TO_NUTS_SPEC/nuts-registry.yaml > generated/registry.gen.go

The generated code requires another dependency

.. code-block:: shell

   go get github.com/labstack/echo

Display help
------------

.. code-block:: shell

   go run main.go help


Run the server
--------------

.. code-block:: shell

   go run main.go
