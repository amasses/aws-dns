AWS DNS resolver
==================

This tool provides a local DNS resolver for AWS servers.

Upon startup, and every 5 minutes (customizable) each AWS profile from the ~/.aws/credentials file is enumerated and each instance is listed.

From this the name tag is queried and if present this is used for the DNS name, otherwise the instance ID is used. The suffix of .aws is used for resolving.

A custom resolver in /etc/resolver is setup to direct all DNS queries for whatever.aws to the local resolver listening on port udp/10053.

This is known to work on OSX (and uses launchd to run the server process) but may also work on Linux systems with a bit of fiddling.

Installing
----------

A valid GoLang environment and make is required to build and install.

Simply run the following commands (note `sudo` is called to install the binary and resolver):

```
make
make install
```

To remove run:

```
make uninstall
```

Errata
------
I cannot guarantee this will work, but all operations performed are read only so rest assured this app won't go and reboot or destroy your EC2 instances. If you are concerned about this you can setup an IAM role which only has read access to the EC2 data and set this up in your credentials file.

Improvements welcome :)
