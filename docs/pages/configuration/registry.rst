.. _nuts-registry-configuration:

Nuts registry configuration
###########################

.. marker-for-readme

Sync modes
==========

The registry supports two modes for updating the internal Db: a file system watcher (``fs``) or downloading from Github (``github``).
When using Github, the registry checks every ``syncInterval`` minutes if anything has changed on Github.
The ``syncAddress`` must point to a tar.gz with the needed registry files included. Github has a nice URL for this.
By default it uses the config in the master branch.
