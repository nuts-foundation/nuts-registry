.. _reverse-proxy-configuration:

Using reverse proxies
#####################

The Nuts network makes heavy use of TLS connections with client certificates (a.k.a. mTLS).
This documentation covers some guidelines and examples when using a reverse proxy for TLS termination and to secure your data endpoint and Nuts endpoints. It only covers HTTP/REST calls and not gRPC.
In the `Examples`_ section, examples are given for the most popular proxies.
This page only covers the client certificate part. Server certificates must be configured as normal, they must be signed by one of the globally accepted CAs (like Let's Encrypt and others).

Trusted certificates
********************

.. warning::

    It is essential to do this part correctly, otherwise the endpoints will be exposed to the outside world!

Trusted roots
=============

The basis for setting up security for mTLS is to get hold of the correct root certificates.
For the internet, the trusted roots are preconfigured by the OS and/or browser.
Root certificates are often published on websites. Check out https://letsencrypt.org/certificates/ for an example.
For the Nuts network, several choices can be made, the matrix below shows a proposal on which CA bundle to trust.

.. note::

    This is subject to change based on peer-reviews

+----------------+----------------------+------------------+--------------+---------------------------+---------------------------+
|                | local                | development      | demo         | test                      | production                |
+================+======================+==================+==============+===========================+===========================+
| nuts endpoints | none/local generated | Nuts dev CA      | nuts demo CA | nuts test CA              | nuts production CA        |
+----------------+----------------------+------------------+--------------+---------------------------+---------------------------+
| data endpoints | none                 | Nuts dev CA      | nuts demo CA | PKIo with allow list      | PKIo with allow list      |
+----------------+----------------------+------------------+--------------+---------------------------+---------------------------+

Nuts endpoints cover all APIs provided by Nuts nodes. Data endpoints are provided by the vendor and cover patient related data.
The **local**, **development** and **demo** CA bundles follow the same convention for both the nuts endpoints and the data endpoints.
Since no real data will be used within these environments, using the different Nuts root certificates can simplify deployment.

For **test** and **production** this is a different story. **test** is a synonym for **acceptance**, thus real patient data is transferred.
A national recognized trusted root in combination with specific trusted certificates gives maximum security.
Configuring the different PKIo CAs in the reverse proxy will filter out most unwanted traffic and makes any traffic traceable to the responsible party.
Using an allow list on top of that will give control to the vendor which certificates to accept.
The allow list will be available from the registry. Only specifically published certificates get added to the allow list.
More on the allow list can be read below. The different PKIo roots can not be obtained from the Nuts APIs.

Everywhere where the term *Nuts X CA* is used, the Nuts foundation acts as the **network authority**.
The network authority role is only performed by a single party (legal entity) and it provides the CA bundles.

The Nuts foundation publishes its CA bundles on https://nuts.nl
A full Nuts CA bundle would consist of a root, a Nuts intermediate and multiple vendor CAs.

The ``/api/mtls/*`` APIs on :ref:`nuts-registry-api` describe how to obtain a list of CAs in various formats.
Most reverse-proxies do not support calling APIs but only file based configuration.
It's up to the vendor to convert the API output to a file and restart the proxy when something has changed.
A single ``truststore.pem`` file with all CAs in PEM format is available in the directory configured by the ``nuts.crypto.fspath`` variable.
This file can be synced or linked to from various configs.

Allow list
==========

All client certificates used for mTLS will be published in the Nuts registry.
This will allow for extracting the list of certificates that will be used on the application level.
Vendors control which certificates from the allow list are passed on to the reverse-proxy,
giving them complete control over which certificates get access.
This means that vendors will have full control over which certificate is granted access and which isn't.

Most reverse-proxies require a configured list of CAs to enable mTLS.
Most of the time, however, they will not be able to configure a list of accepted non-CA certificates.
Specific certificate acceptance must therefore be done at the application level.

The ``/api/mtls/*`` APIs on :ref:`nuts-registry-api` describe how to obtain an allow list in various formats.

Examples
********

Below are configuration examples on how to configure various reverse proxies.

Apache
======

Used apache version: 2.4.43

.. code-block:: xml

    <VirtualHost *:443>
    ServerName proxy-server
      # activate HTTPS server certificate on the reverse proxy
        SSLEngine on
        SSLCertificateFile "<Apache_home>/conf/proxy-server.crt"
        SSLCertificateKeyFile "<Apache_home>/conf/proxy-server.key"
      # activate the client certificate authentication
        SSLCACertificateFile "/etc/nuts/data/truststore.pem"
        SSLVerifyClient require
        SSLVerifyDepth 1
        SSLProxyEngine On
      # initialize the special headers to a blank value to avoid http header forgeries
        RequestHeader set SSL_CLIENT_CERT ""
        <Location /api/fhir>
            # forward certificate for allow list checking
            RequestHeader set SSL_CLIENT_CERT "%{SSL_CLIENT_CERT}s"
            ProxyPass http://localhost:8080
            ProxyPassReverse http://localhost:8080
        </Location>
    </VirtualHost>

HAProxy
=======

.. note::

    help requested on a valid HAProxy example

Nginx
=====

Used nginx version: 1.17.0

.. code-block:: text

    http {
      # only TLS > 1.2 is acceptable
      ssl_protocols TLSv1.2 TLSv1.3;
      ssl_prefer_server_ciphers on;

      access_log /var/log/nginx/access.log;
      error_log /var/log/nginx/error.log;

      # server on port 80 for HTTP -> HTTPS redirect
      server {
        listen 80;
        server_name example.com;
        return 301 https://example.com$request_uri;
      }

      # The HTTPS server, which proxies our requests
      server {
        listen 443 ssl;
        server_name example.com;

        ssl_protocols TLSv1.2 TLSv1.3;
        # server certificate
        ssl_certificate /etc/nginx/ssl/example.com/fullchain.pem;
        ssl_certificate_key /etc/nginx/ssl/example.com/privkey.pem;

        # client certificate,
        # here we use the exported truststore.pem in the case Nuts is
        # running on the same machine
        ssl_trusted_certificate /etc/nuts/data/truststore.pem;
        ssl_verify_client on;

        access_log /var/log/nginx/example.com;

        location /api/fhir {
          proxy_set_header        Host $host;
          proxy_set_header        X-Real-IP $remote_addr;
          proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
          proxy_set_header        X-Forwarded-Proto $scheme;
          # forward certificate for allow list checking
          proxy_set_header        X-Ssl-Client-Cert $ssl_client_cert;

          proxy_pass          http://localhost:8080;
        }
      }
    }
