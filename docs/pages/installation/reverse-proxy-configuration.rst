.. _reverse-proxy-configuration:

Using reverse proxies
#####################

The Nuts network makes heavy use of TLS connections with client certificates (a.k.a. mTLS).
It's therefore highly recommended to use a reverse proxy for TLS termination and to secure your data endpoint and Nuts endpoints.
In the `Examples`_ section, examples are given for the most popular proxies.
The page only covert the client certificate part. Server certificates must be configured as normal, they must be signed by on of the globally accepted CAs.

Trusted certificates
********************

.. warning::

    It is essential to do this part correctly, otherwise the endpoints will be exposed to the outside world!

Trusted roots
=============

The basis for setting up security for mTLS is to get hold of the correct root certificates.
For the internet, the trusted roots are preconfigured by the OS and/or browser.
Root certificates are often published on websites. Check out https://letsencrypt.org/certificates/ for an example.
For the Nuts network, several choices can be made, the matrix below shows a proposal on which roots to trust.

.. note::

    This is subject to change based on peer-reviews

+----------------+----------------------+------------------+-----------+---------------------------+---------------------------+
|                | local                | development      | demo      | test                      | production                |
+================+======================+==================+===========+===========================+===========================+
| nuts endpoints | none/local generated | nuts development | nuts demo | nuts test                 | nuts production           |
+----------------+----------------------+------------------+-----------+---------------------------+---------------------------+
| data endpoints | none                 | nuts development | nuts demo | PKIo with acceptance list | PKIo with acceptance list |
+----------------+----------------------+------------------+-----------+---------------------------+---------------------------+

Nuts endpoints cover all API's provided by Nuts nodes. Data endpoints are provided by the vendor and cover patient related data.
The **local**, **development** and **demo** roots follow the same convention for both the nuts endpoints and the data endpoints.
Since no real data will be used within these environments, using the different Nuts root certificates can simplify deployment.

For **test** and **production** this is a different story.
A national recognized trusted root in combination with specific trusted certificates gives maximum security.
More on the acceptance list can be read below. The different PKIo roots can not be obtained from the Nuts API's.

The different Nuts root certificates will be published on https://nuts.nl
A full Nuts chain would consist of a root, a Nuts intermediate and multiple vendor CAs.

The ``/api/mtls/*`` APIs on :ref:`nuts-registry-api` describe how to obtain a list of CAs in various formats.
Most reverse-proxies do not support calling API's but only file based configuration.
It's up to the vendor to convert the API output to a file and restart the proxy when something has changed.
A single ``truststore.pem`` file with all CAs in PEM format is available in the directory configured by the ``nuts.crypto.fspath`` variable.
This file can be synced or linked to from various configs.

Acceptance list
===============

All client certificates used for mTLS will be published in the Nuts registry.
This will allow for extracting the list of certificates that will be used to the application level.
Vendors could use this information to enable or disable certain certificates on demand.
This means that vendors will have full control over which certificate is granted access and which isn't.

Most reverse-proxies require a configured list of CAs to enable mTLS.
Most of the time, however, they will not be able to configure a list of accepted non-CA certificates.
Specific certificate acceptance must therefore be done at the application level.

The ``/api/mtls/*`` APIs on :ref:`nuts-registry-api` describe how to obtain an acceptance list in various formats.

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
        SSLCertificateChainFile "<Apache_home>/conf/proxy-server-ca.crt"
      # activate the client certificate authentication
        SSLCACertificateFile "/etc/nuts/data/truststore.pem"
        SSLVerifyClient require
        SSLVerifyDepth 1
        SSLProxyEngine On
      # initialize the special headers to a blank value to avoid http header forgeries
        RequestHeader set SSL_CLIENT_CERT ""
        <Location /api/fhir>
            # forward certificate for acceptance list checking
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
        # the client certificate is optional, so for example status endpoints can be accessed
        ssl_verify_client optional;

        access_log /var/log/nginx/example.com;

        location /api/fhir {
          # make the certificate non-optional for this path
          # if the client-side certificate failed to authenticate, show a 403
          # message to the client
          if ($ssl_client_verify != SUCCESS) {
            return 403;
          }

          proxy_set_header        Host $host;
          proxy_set_header        X-Real-IP $remote_addr;
          proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
          proxy_set_header        X-Forwarded-Proto $scheme;
          # forward certificate for acceptance list checking
          proxy_set_header        X-Ssl-Client-Cert $ssl_client_cert;

          proxy_pass          http://localhost:8080;
        }
      }
    }
