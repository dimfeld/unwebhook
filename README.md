unwebhook
=============

Webhook server for Gitlab and Github to run arbitrary commands based on events. Currently under development.

The server supports multiple hooks simultaneously, exposed at customizable URLs.

Note that I'm only using this regularly with Gitlab over a local network for push events, handling no more than a couple of requests each day. This means that it has not been well-tested in anything remotely close to a production environment. But I'm glad to accept pull requests or try to fix any issues you find.

## Usage

The program reads information from a configuration format, described below. Configuration can be sourced from multiple files and directories. 

The first command line argument is the main configuration file, whicn contains hook definitions and other server-wide configuration. Additional configuration files can contain only hooks.

The main configuration file may also be specified in the `UNWEBHOOK_CONFFILE` environment variable. In this case, all command line arguments are limited to containing only hooks.

If no configuration files are specified, the program looks for a file with of the format *executable-name*.conf.

```shell
# Read configuration from unwebhook.conf
% unwebhook

# Read main configuration from main.cfg, and load hooks from all files in webhook-shell.conf.d
% unwebhook main.cfg unwebhook.conf.d

# Read main configuration from main.cfg, and load hooks from all files and directories on the command line
% UNWEBHOOK_CONFFILE=main.cfg unwebhook global-hooks.conf hooks.d

# Override listen address from environment
% UNWEBHOOK_LISTENADDRESS=:8080 unwebhook
```

## Configuration
All configuration is in the [TOML](https://github.com/mojombo/toml) format. 

### Server Configuration
These variables are read from the main configuration file and determine server-wide behavior. They may also be overriden by environment variables of the format `UNWEBHOOK_VARIABLENAME`

#### ListenAddress
The address and port on which to listen. If none is provided, the default `:80` is used.

```
ListenAddress = 127.0.0.1:8080
```

#### CommandTimeout
The maximum time, in seconds, that any single command is allowed to run.

```
CommandTimeout = 5
```

#### AcceptIp
A list of IP addresses from which to accept requests. Requests from non-allowed IPs are logged and ignored.

If not specified, requests are allowed from any IP address.

```
AcceptIp = [ "172.17.0.1", "192.168.1.65" ]
```

#### Secret
A string used as a key to calculate an HMAC digest of the request body. Requests that don't have a matching
digest will be ignored. Note that Gitlab does not support this feature.

If specified, this overrides any server-wide secret. If a secret is present in the server-wide configuration, it can be disabled for this hook by setting the hook's secret to "none".

#### LogDir
The directory of the log file. If not given, the default is the current directory. This can also be specified on the command line using the -log_dir command-line option.

```
LogDir = "/var/log/unwebhook"
```

#### HookPaths
An optional list of files and directories, from which the server will load hooks. These paths are in addition to any hooks defined in the main configuration file or additional command line arguments.

```
HookPaths = [ "/etc/unwebhook/conf.d", "/etc/unwebhook/hooks.conf" ]
```

#### Hook
An optional list of Hook objects. When given in the server configuration file, this looks identical to a list of hooks as described in the Hook Configuration section below. Due to limitations of the TOML format, any hook definitions in the main configuration file must be at the end.

### Hook Configuration

#### Url

#### Dir

#### Env

#### PerCommit

#### AllowEvent

#### Secret
A string used as a key to calculate an HMAC digest of the request body. Requests that don't have a matching
digest will be ignored. Note that Gitlab does not support this feature.

If specified, this overrides any secret from the server-wide configuration. If a secret is present in the server-wide configuration, it can be disabled for this hook by setting the hook's secret to "none".

#### Timeout

#### Commands

#### Command Templates

### Sample Configuration
