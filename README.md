Riker
-----
gRPC distributed Chat bot. Riker is the server that other apps register commands
with. We call the command providers `redshirts`. Riker handles the complexity of
interfacing with a chat provider such as slack. Redhirts communicate with Riker
using gRPC. Authentication between Riker and redshirts is mutual TLS. Redshirts
register a command, and users/groups that are authorized to use it.


Getting Started
===============

You need these things to get up and running
* Bot token from slack
* API token from slack
* TLS cert+key for riker server
* Optional: the CA cert for the CA that was used to make your .pem. If you used
  A known public CA then you won't need this.

First you will have to create a slack bot in the slack admin UI. Whatever you
name the bot in slack you will need to pass to Riker as the name field. Make
sure you have a TLS cert+key in .pem format for riker.

### Get riker
you can get a release from there releases tab.

### Configure Riker
#### Environment variables
Riker supports every commandline flag being specified as an environment variable.
To use environment variables prefix any flag with `RIKER` and convert any dashes
to underscores. For example `--bind-address` would be specified in the environment
as `RIKER_BIND_ADDRESS`

#### Config files
Riker supports a config file in any of HCL, TOML, YAML, JSON or Java properties
formats. Config files are looked for in the current directory `./.riker.<ext>`
as well as in `~/.riker.<ext>`. This simple config should get you up and running.


```
# Place this in  ~/.riker.yaml
---
bot-name: riker
debug: true
json-log: true
bind-address: localhost:6000

# change these to your tokens
bot-token: yourbottoken
api-token: yourapitoken

# use your own certs here
tls-cert: ~/path/to/riker.pem
ca-cert: ~/path/to/ca.crt

# Allowed-ou are a further restriction on the TLS for connecting redshirts.
# only clients connecting with one of these OUs will be allowed to register a
# command.
allowed-ou:
  - riker
  - riker-redshirt

# healthz options
healthz-port: 8080
healthz-address: 0.0.0.0
```

### Running Riker
Assuming you have a config file you can run Riker with the slack chat provider:

```
./riker slackbot
```

### Monitoring
Cuurently, Riker has a "naive" health monitoring mechanism in place to determine if it's running, the Kube way. It exposes two HTTP endpoints:
- /healthz -> kube readiness endpoint
- /liveness -> kube liveness endpoint

TODO: integrate proper health monitoring of Riker via something like [grpc-health-probe](https://github.com/grpc-ecosystem/grpc-health-probe)

### Kube / Helm chart
TODO

Chat Providers
==============
### Slack
Slack is what we use at pantheon, and what riker first was built against. Currently
We don't have plans to implement other backends, but PR's are welcome.

#### Authenticating riker to slack:
There are 3 things you need (4 if you run your own CA)
- RIKER_API_TOKEN: Used by riker to authenticate to the Slack API
- RIKER_BOT_TOKEN: Used by riker to login to Slack "chat" as a bot.

### Local Terminal
To help develop redshrits we implemented a local mode chat server that disables
gRPC authentication. This is meant only for local development. It can also be a
way to play with Riker.

You can set the nickname of the terminal mode user and the groups that the user belongs to using
the command line options --nickname and --groups. Groups should be a comma seperated list of group
names.

Components
==========
- Riker: The server and chat gateway. Enforces authentication and authorization
  of all commands.
- redshirt: The client and command provider. A "microbot" registers to riker
  with credentals

Building From Source
====================

Riker should be build using go1.9+. Riker uses the `dep` tool for handling
dependencies, but the deps are not special. You should be able to build with
`go get`.
```
go get -u github.com/pantheon-systems/riker
```

If that doesn't work you should check out the source and use dep ensure.
```
git clone https://github.com/pantheon-systems/riker.git
```

Then you just need to run make to setup deps and build
```
$ cd riker
$ make deps build
```

Developing Redshirts
====================
TODO: add alot more info  here, but this is the gist
- Run riker in terminal mode.
- Connect shirt without auth.
- proffit.
