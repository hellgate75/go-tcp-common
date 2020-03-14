<p align="center">
<image width="150" height="50" src="images/kube-go.png"></image>&nbsp;
<image width="260" height="410" src="images/golang-logo.png">
&nbsp;<image width="130" height="50" src="images/tls-logo.png"></image>
</p><br/>
<br/>

# Go Tcp Client / Server Common library

Code used for computation, logging, thread pool manageemnt, and more...

## Affected Repositories


Here list of linked repositories:

[Server Repository](https://github.com/hellgate75/go-tcp-server)

[Client Repository](https://github.com/hellgate75/go-tcp-client)

[Modules Repository](https://github.com/hellgate75/go-tcp-modules)

[Go Deploy](https://github.com/hellgate75/go-deploy)

[Go Deploy Modules](https://github.com/hellgate75/go-deploy-modules)

[Go Deploy Clients](https://github.com/hellgate75/go-deploy-clients)

<br/>


## Provided libraries:

* [common](/common/common.go) - Common model entities

* [io -> files tools](/io/files.go) - File operations

* [io -> parsers tools](/io/parsers.go) - Json, Yaml, Xml parsing libraries

* [io/streams -> data stream](/io/streams/data-stream.go) - Fluent DataStream model

* [log](/log/logger.go) - System Logger

* [net/api/client](/net/api/client/client.go) - Api Client (TLS/No TLS) declarations and implementation

* [net/api/common](/net/api/common/common.go) - Api Client Commmon Models

* [net/api/server](/net/api/server/server.go) - Api Server (TLS/No TLS) declarations and implementation

* [net/common](/net/common/servers.go) - Common Net interfaces

* [net/rest/common](/net/rest/common/net.go) - Common Net Rest interfaces

* [net/rest/tls/client](/net/rest/tls/client/client.go) - Rest TLS Client (TLS/No TLS) declarations

* [net/rest/tls/client -> impl](/net/rest/tls/client/client-funcs.go) - Rest TLS Client (TLS/No TLS) implementation

* [net/rest/tls/server](/net/rest/tls/server/server.go) - Rest TLS Server (TLS/No TLS) declarations

* [net/rest/tls/server -> impl](/net/rest/tls/server/server-funcs.go) - Rest TLS Server (TLS/No TLS) implementation

* [net/rest/tls -> clients](/net/rest/tls/clients.go) - Rest TLS Clients export interfaces

* [net/rest/tls -> servers](/net/rest/tls/servers.go) - Rest TLS Servers export interfaces

* [pool](/pool/threads.go) - Thread Pool component and related interfaces and sub-components

<br/>

## Available samples

Samples are available for:

* [SSL/TLS Rest sample Server / Client lifecycle](/samples/rest/tls/server-client.go)

* [API SSL/TLS security sample Server / Client lifecycle](/samples/api/tls/server-client.go)


Enjoy the experience.

## License

The library is licensed with [LGPL v. 3.0](/LICENSE) clauses, with prior authorization of author before any production or commercial use. Use of this library or any extension is prohibited due to high risk of damages due to improper use. No warranty is provided for improper or unauthorized use of this library or any implementation.

Any request can be prompted to the author [Fabrizio Torelli](https://www.linkedin.com/in/fabriziotorelli) at the following email address:

[hellgate75@gmail.com](mailto:hellgate75@gmail.com)

