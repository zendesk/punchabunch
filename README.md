Punchabunch: A configurable SSH local forwarding proxy
======================================================

Punchabunch is a simple tool for setting up multiple SSH local forwarding
proxies without a lot of configuration.  It simplifies the task of setting
up a large number of connections that must be proxied through multiple
SSH servers.

Punchabunch doesn't do anything that `ssh(1)` and a well-written
wrapper script and `ssh_config(5)` generator couldn't do -- but it saves
the user the burden of doing so.  And it comes in a neatly packaged binary
with a very simple configuration format.

Build instructions
------------------
1. Install Go, either from the [official distribution](https://golang.org/dl/) or via
   [Homebrew](http://brew.sh).
1. Run `go get github.com/zendesk/punchabunch`

Usage
-----
```
punchabunch [-config config_file]

  -config string
    	Path to configuration file (default "config.toml")
```


Configuration
-------------

Punchabunch uses a [TOML](https://github.com/toml-lang/toml) configuration
file to describe the desired set of SSH servers and forwarding configurations.
It's as simple as a set of entries that look like this:

```
[app]
bastion = "bastion.example.com"
listen = "12345"
forward = "app.internal.example.com:80"

[db]
bastion = "bastion.example.com"
listen = "12345"
forward = "db.internal.example.com:3306"
```

Each entry must be preceded by an arbitrary `[name]` header (including the
brackets).

The keys for each entry are as follows:

* `bastion`: The bastion host (server) to proxy the connection through.
  The value must be a string of the format `host[:port]`.  If the `port`
  value is not specified, port 22 will be used.

* `listen`: The local port number to listen to for requests.  The value
  must be a string of the format `[interface:]port`.  By default, the
  proxy will bind to the local IPv4 loopback address (127.0.0.1). If
  you prefer IPv6, set this value to `[::1]:<port>`.

* `forward`: The destination host:port pair that the `listen` port
  will forward incoming requests to.  It must be a string of the format
  `host:port`.

You can have as many of these entries as your resources can accommodate.
Punchabunch will start as many SSH sessions as required, in parallel.  Only one
SSH connection will be established per bastion host, even if multiple
forwarders are configured across it.

Authentication
--------------
Currently, Punchabunch requires an active SSH Agent (see `ssh-agent(1)`)
to operate, and it cannot be configured with other sources of
authentication information.   If you have a working agent, no
configuration is needed -- Punchabunch will automatically locate its
local socket.

In the future, Punchabunch may be able to directly read SSH private keys.

Authors
-------
* Michael S. Fischer, [mfischer@zendesk.com](mailto:mfischer@zendesk.com)

Reporting bugs
--------------
Please report bugs or other issues at https://github.com/zendesk/punchabunch/issues.
