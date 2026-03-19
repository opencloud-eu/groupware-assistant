# Groupware Assistant

A command-line tool to assist with developing and testing the OpenCloud Groupware, by creating realistic-ish looking data of various kinds.

It is a JMAP client that directly interacts with a JMAP server, and not with the OpenCloud Groupware itself.

## Authentication

The application currently only supports authenticating through basic authentication credentials, which means that the JMAP server needs to allow that.

The global command-line options `--username`, `--password` and `--account-id` provide the means to select how to authenticate and then which account to use for that user, respectively.

By default, it will use `alan` with the password `demo` for authentication, since that is typically the default one used in development instances of OpenCloud.

When the account to use is not specified using `--account-id` (or `-A`), the matching primary account will be pulled from the JMAP session and used.

## Server

The `--url` option may be used to specify the HTTP URL of the JMAP server to use when performing operations.

It defaults to `https://stalwart.opencloud.test`, which is the URL to use for the `devtools/deployments/opencloud_full` development deployment.


