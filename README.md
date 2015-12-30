h2d - A Simple HTTP/2 Command-Line Client
-----------------------------------------

`h2d` is a simple HTTP/2 command-line client, like `curl`.

While `curl` terminates after each request/response cycle, `h2d` runs a background process to keep connections open.
That way, `h2d` may receive asynchronous [push](https://httpwg.github.io/specs/rfc7540.html#PushResources) messages from the server.

`h2d` is currently in a very early stage. The best way to learn about it is to read the blog posts on [unrestful.io](http://unrestful.io).

Screenshots
-----------

![h2d dump](doc/h2d-dump.png)

![h2d command line](doc/h2d-cmdline.png)

Basic Usage
-----------

```bash
h2d start &
h2d connect http2.akamai.com
h2d get /index.html
h2d stop
```

Command Overview
----------------

For a complete list of available commands, run `h2d --help`.

* `h2d start [options]` Start the h2d process. The h2d process must be started before running any other command.
* `h2d connect [options] <host>:<port>` Connect to a server using https
* `h2d disconnect` Disconnect from server
* `h2d get [options] <path>` Perform a GET request
* `h2d post [options] <path>` Perform a POST request
* `h2d set <header-name> <header-value>` Set a header. The header will be valid for all subsequent requests.
* `h2d unset <header-name> [<header-value>]` Undo 'h2d set'.
* `h2d pid` Show the process id of the h2d process.
* `h2d push-list` List responses that are available as push promises.
* `h2d stop` Stop the h2d process
* `h2d wiretap <localhost:port> <remotehost:port>` Listen on localhost:port and forward all traffic to remotehost:port.

How to Download and Run
-----------------------

Binary releases are available on the [GitHub Releases](https://github.com/rmohid/h2d/releases).

1. Download the latest release ZIP file: [h2d-v0.0.8.zip](https://github.com/rmohid/h2d/releases/download/v0.0.8/h2d-v0.0.8.zip)
2. Extract the ZIP file
3. Find the executable for your system in the `bin` folder:
  * Linux: `h2d_linux_amd64`
  * OS X: `h2d_darwin_amd64`
  * Windows: `h2d_windows_amd64.exe`
4. Rename that executable to `h2d`, or `h2d.exe` on Windows
5. Move the executable into a folder on your PATH.

How to Build from Source
------------------------

`h2d` is developed with [go 1.5.1](https://golang.org/dl/). The external dependencies are located in the `vendor` folder,
in order to load these dependencies, you must enable the
[Go 1.5 Vendor Handling](http://engineeredweb.com/blog/2015/go-1.5-vendor-handling/)
by setting the environment variable `GO15VENDOREXPERIMENT` to `1`.

The following command will download, compile, and install `h2d`:

```bash
export GO15VENDOREXPERIMENT=1
go get github.com/rmohid/h2d
```

Related Work
------------

`h2d` uses parts of Brad Fitzpatrick's [HTTP/2 support for Go](https://github.com/bradfitz/http2). There is an HTTP/2 console debugger included in [bradfitz/http2](https://github.com/bradfitz/http2), but just like `h2d`, it is currently only a quick few hour hack, so it is hard to tell if they aim at the same kind of tool.

LICENSE
-------

`h2d` is licensed under the [Apache License, Version 2.0](LICENSE).

`h2d` is implemented in [Go](https://golang.org) and uses Go's [standard library](https://golang.org/pkg/#stdlib), which is licensed under [Google's Go license](https://code.google.com/p/go/source/browse/LICENSE), which is a variant of the [BSD License](https://en.wikipedia.org/wiki/BSD_licenses).

The following 3rd party libraries are used:

  * `golang.org/x/net/http2/hpack` implements the [Header Compression for HTTP/2 (HPACK)](https://httpwg.github.io/specs/rfc7541.html). The library is licensed  [under the terms of Go itself](https://github.com/bradfitz/http2/blob/master/LICENSE).
  * `github.com/fatih/color` implements the color output used in `h2d start --dump`. The library is licensed under an [MIT License](https://github.com/fatih/color/blob/master/LICENSE.md).
