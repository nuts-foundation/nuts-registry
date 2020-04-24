.. _nuts-registry-technical:

Nuts Registry Data Model
########################

The Nuts Registry is event-sourced, meaning its data is stored as events which are replayed to determine the current state.
The event store is append-only which makes it easier to migrate to a distributed architecture.

An event contains the following fields:

=====     ======  =====  =============================================  ========
Field     Type    Since  Description                                    Example
=====     ======  =====  =============================================  ========
issuedAt  string  v0     Time at which the event was issued
type      string  v0     Type of the payload                            RegisterVendorEvent
jws       string  v0     Optional. JWS-encoded signature of the event
payload   object  v0     Actual payload of the event                    ``{"Message": "Hello, World!"}``
version   int     v1     Version of the event                           1
ref       string  v1     Hex-encoded reference to this event            67f732c4a421c8d7e097dfa55a27b67b4c5fbd9e
prev      string  v1     Hex-encoded reference to the previous event    67f732c4a421c8d7e097dfa55a27b67b4c5fbd9e
=====     ======  =====  =============================================  ========

It contains 2 sections, its headers (``issuedAt``, ``type``, `jws`, ``version``, ``ref``, ``prev``) and a human-readable copy of the content of event (``payload``).
We call it a copy because the authenticated content of the event is found encoded inside the `jws` field. The ``payload``
field is there for ease of use but will be removed in a future version and thus should not be relied on.
The ``type`` field defines the type of the event (e.g. ``RegisterVendorEvent``), indicating how to interpret the payload.

Signing
*******

The events are signed to assure authenticity (proof that the event was actually created by the expected entity) and
integrity (proof that the contents weren't altered by an attacker). The entity that 'owns' the data the event generates
must be the one signing the event (see the table below).

Signature format
================

The signatures are in the JWS (`JSON Web Signature <https://tools.ietf.org/html/rfc7515>`_) format. Since every key should be associated to a known entity,
the JWS will contain an X.509 certificate (in the ``x5c`` field) describing the entity owning the key. The algorithm
used for constructing the JWS is RS256 (RSA with SHA-256 hashing function).

The JWS also contains the actual payload of the event.

Validation
==========

To validate an event signature, the following checks must be performed:

1. Is the JWS parsable?
2. Does the JWS contain an X.509 certificate chain (in the ``x5c`` field)?
3. Is the certificate meant for signing (key usage must contain ``digitalSignature``)
4. Is the certificate (extracted from the chain) trusted?
5. Was the certificate valid at the time of signing (``issuedAt``)?
6. Is the owning entity of the certificate (e.g. a vendor or organization) the one we expected to sign the certificate (see *Owner check* in the table below)?
7. Is the JWS signed using the RS256 algorithm?
8. Is the used RSA key of sufficient length (>=2048 bits)?
9. Is the JWS signed with the private key belonging to the public key in the certificate?

When an event payload containing a CA certificate is successfully validated, it should be added to the node's trust store so that
future events which are signed using certificates issued by the (CA) certificate can be validated.

======================  ============  ===========
Event                   Signer        Owner check
======================  ============  ===========
RegisterVendor          Vendor        ``Event.Payload.Vendor == Certificate.SubjectAltName[Vendor]``
VendorClaim             Vendor        ``Event.Payload.Vendor == Certificate.SubjectAltName[Vendor]``
RegisterEndpoint        Organization  ``Event.Payload.Organization == Certificate.SubjectAltName[Organization]``
======================  ============  ===========

.. note::
    The Nuts Foundation will act as Root Certificate Authority so that intermediates are issued by an entity which is trusted
    by all participating parties. However, this Root Certificate Authority isn't operational at the time of writing so
    vendors are expected to self-sign their own CA certificates in the meantime.
    This means when validating a ``RegisterVendor`` event the certificate which signed the JWS will be self-signed and
    thus can't be validated. **This is the only case** where an unvalidated certificate should be added to the trust store.

Versioning
**********

Events have a ``version`` field indicating the version of the data structure. New versions might introduce new fields or
change the datatype of existing fields (although this change must be backwards compatible.

===========  ==================================================
Version      Change
===========  ==================================================
0            Any version before introduction of ``version``
1            ``version``, ``ref`` and ``prev`` added
2 (planned)  JWS signed payload is canonicalized before hashing
===========  ==================================================