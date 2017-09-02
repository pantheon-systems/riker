- Could allow redshirts to register on events as well as commands, or commands to events.
  in this mode we could passthrough  the slack event to the redshirt..

- [ ] factor out the gRPC authOU and related funcs into our go-certauth lib? They're similar in spirit

---------------------------------------------------------------------------------------------
https://getpantheon.atlassian.net/browse/IO-2612
"CSE.py actions available securely in Slack"

Done when:
- [ ] Enable tasks for specific CSE users
- [ ] Document Chatops, Train CSE
- [ ] Update CSE workflows wiki pages to
- [ ] Remove CSE access from critical production systems

Tasks:
- [ ] Deploy Riker, and redshirt army
- [x] Figure out auth riker<-->redshirt
    we'll use mTLS
- [ ] document how to use the language-agnostic proxy/lieutenant/interface/adapter thingy so that other eng team's can build redshirts
- [ ] riker-proxy/lieutenant wrapper/adapter app
- [x] fix bug where, when using private chat, you get both a direct response and threaded response from riker
- [ ] bug: riker does not realize when a redshirt disconnects. It should remove disconnected clients from the b.redshirts map.
-   Upon reconnect, any new user/group perms will not take effect which can cause problems.
  - (joe): for now i worked around this in cli-wrapper by setting ForcedRegistation: true, since most use cases are singleton redshirts


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


---------------------------------------------------------------------------------------------
README / docs

riker
=====

terminology:
- riker: server .. the gateway/orchestrator .. message-bus for humans and robots.. gateway between slack and redshirts..
         enforces authentication and authorization of all commands ..
- redshirt: client .. a microbot .. provider of a command namespace


Credentials:
- SLACK_TOKEN: Used by riker to authenticate to the Slack API
- SLACK_BOT_TOKEN: Used by riker to login to Slack "chat" as a bot.

