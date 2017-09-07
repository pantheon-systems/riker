- Could allow redshirts to register on events as well as commands, or commands to events.
  in this mode we could passthrough  the slack event to the redshirt..

TODO:

## Long term
- [ ] Offline testing - allowing to test without slack.
- [ ] factor out the gRPC authOU and related funcs into our go-certauth lib? They're similar in spirit
- [ ] OSS the project
	* better docs
	* in pantheon or own org ?
	* License? (APL2 likely because we are using some APL code)
- [ ] External state
- [ ] Consenus based redshirt registration (versioned capabilities)


---------------------------------------------------------------------------------------------
https://getpantheon.atlassian.net/browse/IO-2612
"CSE.py actions available securely in Slack"

## Done when:
- [ ] Enable tasks for specific CSE users
- [ ] Document Chatops, Train CSE
- [ ] Update CSE workflows wiki pages to
- [ ] Remove CSE access from critical production systems

## Short Term:
### Riker Server
- [x] bug: riker does not realize when a redshirt disconnects. It should remove disconnected clients from the b.redshirts map.
-   Upon reconnect, any new user/group perms will not take effect which can cause problems.
  - (joe): for now i worked around this in cli-wrapper by setting ForcedRegistation: true, since most use cases are singleton redshirts
	- Jesse: calling this done per 96cb864. the map doesn't track clients, and the agreed upon last registration wins for updating the capabilities.
- [x] Remove forced registration, and use simple always-apply approach for now.
	- Jesse: done in 96cb864
- [ ] Cobrafication /  config file
- [ ] Deploy Riker, and redshirt army
- [x] Figure out auth riker<-->redshirt
    we'll use mTLS
- [x] fix bug where, when using private chat, you get both a direct response and threaded response from riker

### CLI wrapper
- [ ] positional command vs flag
- [x] Break CLI wrapper redshirt to own repo.
- [ ] document how to use so other eng team's can build redshirts
- [ ] cse.py MVP
  - [ ] TODO: break this down into more todos...

### riker-proxy/lieutenant
```
redshirt-proxy \
  -cert foo.pem \
  -namespace "cse" \
  -description="foo bar blah baz" \
  -usage "hrm.maybe this comes from a file" \
  -exec "./my-bot"
```
- uses stdin/out/err to bridge simple app to slack/riker
- should be available as docker container + linux-amd64 binary on github releases? - to make it easy to integrate into
  anyone's deployment


NOTES:
------

## RedShirt capability registration -
Goals:
* make it simple for redshirt writers
* Support multiple shirts upgrading capabilities
* Support roll back
* Support Singleton/Mutex redshirts
* Anti-Flapping

Ideas
- versioned registration
- Auto versioned registration(riker sums the registrations) and applies 'new' ones
- Consensus based. Only apply after some quorum is reached
- Always apply - last writer always wins

Break authentication out from registration for this problem?

---------------------------------------------------------------------------------------------
README / docs

Riker
=====

terminology:
- riker: server .. the gateway/orchestrator .. message-bus for humans and robots.. gateway between slack and redshirts..
         enforces authentication and authorization of all commands ..
- redshirt: client .. a microbot .. provider of a command namespace


Credentials:
- SLACK_TOKEN: Used by riker to authenticate to the Slack API
- SLACK_BOT_TOKEN: Used by riker to login to Slack "chat" as a bot.

