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

.. include:: docs/pages/development/registry.rst
    :start-after: .. marker-for-readme

Configuration
*************

.. include:: docs/pages/configuration/registry.rst
    :start-after: .. marker-for-readme

Parameters
==========

The following configuration parameters are available for the registry:

.. include:: README_options.rst

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
