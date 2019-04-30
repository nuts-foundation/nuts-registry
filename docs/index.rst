.. _nuts-registry:

Nuts Registry
=============

Distributed registry for storing and querying health care organizations their vendors and technical endpoints.
The primary goal of the registry is to translate *well-known* names to technical endpoints.
When a patient grants consent to a health care organization,it does so by name.
This consent has to be copied to the consent registries of that organization.

The registry will also hold the public keys for health care organizations.

In a future version, the registry will also store the identities of those who are authorized to acknowledge a working relationship between a legal entity (the organization) and a user (attribute based identity).

The technical specification of the Nuts registry api's can be found at: :ref:`nuts-node-rpc-registry`

Back to main documentation: :ref:`nuts-documentation`

.. toctree::
   :maxdepth: 2
   :caption: Contents:
   :glob:

   pages/*