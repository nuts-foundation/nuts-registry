nuts-registry
=============

Distributed registry for storing and querying healthcare care providers their vendors and technical endpoints.

.. inclusion-marker-for-contribution

To generate the Server stub install some dependencies:

.. code-block:: shell

   go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen

Then run

.. code-block:: shell

   oapi-codegen PATH_TO_NUTS_SPEC/nuts-registry.yaml > registry/registry.gen.go

The generated code requires another dependency

.. code-block:: shell

   go get github.com/labstack/ech0

Run the server
--------------

.. code-block:: shell

   go run main.go