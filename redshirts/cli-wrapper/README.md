riker cli-wrapper
=================

A wrapper for converting any app into a Redshirt bot

The riker cli-wrapper provides a method for making any command line app or script
accessible as a ChatOps bot via the Riker bot framework.

The wrapper connects to an instance of Riker, registers a command namespace and
a list of authorized users or groups, and then executes a command with arguments
provided by the chat user.

In Riker terminology this is a 'redshirt', or a "micro bot" that implements some
specific functionality.

Usage
-----

Example "echo server":

```shell
	cli-wrapper \
		-addr riker:6000 \
		-cert echo.pem \
		-namespace "echo" \
		-description="echo server" \
		--groups "infra" \
		-usage "echo <msg>: replies with <msg>" \
		-command "/bin/echo"
```

TODO: document the env vars, since this might be more common way to exec in kube?

TODO: example deploy to kubernetes (yaml ..)
