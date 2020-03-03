.. _nuts-registry-technical:

Nuts Registry Data Model
########################

The Nuts Registry is event-sourced, meaning its data is stored as events which are replayed to determine the current state.
The event store is append-only which makes it easier to migrate to a distributed architecture.

An event is structured as follows:

.. code-block::

    Event:
      IssuedAt (timestamp)
      Type (string enum)
      Signature (JWS)
      Payload (object)

It contains 2 sections, its headers (*IssuedAt*, *Type* and *Signature*) and a human-readable copy of the content of event (*Payload*).
We call it a copy because the authenticated content of the event is found encoded inside the *Signature* field. The *Payload*
field is there for ease of use but will be removed in a future version and thus should not be relied on.
The *Type* field defines the type of the event (e.g. "RegisterVendorEvent"), indicating how to interpret the payload.

Signing
*******

The events are signed to assure authenticity (proof that the event was actually created by the expected entity) and
integrity (proof that the contents weren't altered by an attacker). The entity that 'owns' the data the event generates
must be the one signing the event (see the table below).

Signature format
================

The signatures are in the JWS (`JSON Web Signature <https://tools.ietf.org/html/rfc7515>`_) format. Since every key should be associated to a known entity,
the JWS will contain an X.509 certificate (in the *x5c* field) describing the entity owning the key. The algorithm
used for constructing the JWS is RS256 (RSA with SHA-256 hashing function).

The JWS also contains the actual payload of the event.

Validation
==========

To validate an event signature, the following checks must be performed:

1. Is the JWS parsable?
2. Does the JWS contain an X.509 certificate chain (in the *x5c* field)?
3. Is the certificate meant for signing (key usage must contain *digitalSignature*)
4. Is the certificate (extracted from the chain) trusted?
5. Was the certificate valid at the time of signing (*IssuedAt*)?
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
    This means when validating a *RegisterVendor* event the certificate which signed the JWS will be self-signed and
    thus can't be validated. **This is the only case** where an unvalidated certificate should be added to the trust store.