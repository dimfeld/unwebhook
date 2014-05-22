unwebhook
=============

Need some custom functionality in response to repository changes, but don't want to go through the effort of writing a webhook? Unwebhook makes it easy, using a simple configuration format to run any command you want in response to an event.

This hasn't been used in anything approaching a production environment, but it's pretty simple and light on resources.

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

The maximum time, in seconds, that any single command is allowed to run. The default value is 5.

```
CommandTimeout = 5
```

#### AcceptIps

A list of IP addresses from which to accept requests. Requests from non-allowed IPs are logged and ignored.

If not specified, requests are allowed from any IP address.

```
AcceptIps = [ "172.17.0.1", "192.168.1.0/24" ]
```

#### Secret

A string used as a key to calculate an HMAC digest of the request body. Requests that don't have a matching
digest will be ignored. Note that Gitlab does not support this feature.

If specified, this overrides any server-wide secret. If a secret is present in the server-wide configuration, it can be disabled for this hook by setting the hook's secret to "none".

The use of a secret or AcceptIps is highly recommended, since it can protect against malicious data being plugged into your commands.

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
The path at which the server exposes this hook. This does not include the server name, and must start with a slash.

```
Url = "/webhook/"
or
Url = "/webhook/:repo/hook"
```

A path element starting with a colon is a wildcard, which will match on any text in the given path element. In the example above, if the URL configured in GitHub is `/webhook/abc/hook`, then the `repo` variable will be set to `abc`. 

Variables captured in this way are accessible in commands using the syntax `{{ .urlparams.ElementName }}`. In the example here, `{{ .urlparams.abc }}` would be replaced by the text `abc`. See below for more details on the substitution system.

#### Dir
The working directory from which to run the command.

```
Dir = "/home/user/repos"
```

#### Env
A list environment variables, formatted as `key=value` that should replace the current environment. If not given, the current environment is passed into the commands.

```
Env = [ "GONUMPROCS=1", "USER=abc" ]
```

#### PerCommit
If this value is `true`, the hook will be run once for each commit in a push event, with the current commit exposed in the templating system as `.commit`.  A hook comfigured like this will not run anything for an event with no commits.

If the value is `false`, the hook is run once per event.

```
PerCommit = true
```

#### AllowEvent 

A list of events that this hook is allowed to handle. If an event's type is not in this list, it is ignored.

This data is drawn from the `X-Github-Event` HTTP header field for Github events, and from the `object-kind` JSON field for Gitlab events. Gitlab events without an `object-kind` are given the value `push`.

```
AllowEvent = [ "push", "commit_comment" ]
```

#### Secret 

A string used as a key to calculate an HMAC digest of the request body. Requests that don't have a matching
digest will be ignored. Note that Gitlab does not support this feature.

If specified, this overrides any secret from the server-wide configuration. If a secret is present in the server-wide configuration, it can be disabled for this hook by setting the hook's secret to "none".

```
Secret = "abcd"
```

#### Timeout

Overrides the server-wide Timeout setting. Any one command that runs longer than this value, in seconds, will be killed.

```
Timeout = 20
```

#### Commands
A list of commands and arguments to be executed when this hook runs. For each command, the first list item is the executable and the remaining list items are the arguments to that executable. $PATH lookups are performed automatically if no directory is given.

```
Commands = [
  [ "echo", "Received commit, type = {{ .type }}, repo = {{.repository.name}}"  ],
  [ "echo", "                 message = {{.commit.message}}" ]
]
```

### Command Templates
Each string in the  `Dir`, `Env`, and `Command` fields is a template, which can substitute data from the event's payload. Any data in the event's JSON is accessible. The template syntax is provided by Go's `text/template` package, so [full documentation](https://godoc.org/text/template) can be found there.

A template substitution is delimited by `{{ double brackets }}`. Generally, the first character inside brackets will be a `.`, which represents the entire JSON payload.

The template system also provides functions which can transform a particular item in some way. In addition to Go's built-in functions, unwebhook provides a `json` function to print an item and any subitems in JSON format.

The examples below do not represent all of the functions available, but are some of the more useful ones in this context.

```
The repository data, in JSON format.
{{ json .repository }}

The number of commits in the event.
{{ len .commits }}

The first commit, translated into JSON. Note the pipe, which takes the resultof `index .commits 0` and sends it to the JSON function.
{{ index .commits 0 | json }}

The same as above, expressed with nested function calls instead of pipes.
{{ json (index .commits 0) }}

The timestamp of the first commit in the event.
{{ (index .commits 0).timestamp }}

The name of the repository's owner.
{{ .repository.owner.name }}
```

### Environment Variables
In addition to the templating system, the `Dir`, `Env`, and `Commands` members may have environment variables substituted using standard shell syntax such as `Dir="${HOME}/repos"`. The environment variables are taken from the environment in which the server is running, not the environment that may be defined by an `Env` list.

### Sample Configuration

Some sample configuration files can be found in the `conf` directory of this repository. Below is a simple one.

```
ListenAddress = ":8090"
CommandTimeout = 4
LogDir = "/var/log/unwebhook"
HookPaths = [ "/etc/unwebhook.d" ]
AcceptIps = [ "192.30.252.0/24" ]
Secret = "abbadada"

[[Hook]]
Url = "/sync-to-server"
Dir = "$HOME/repos/{{ .repository.name }}"
Env = [ "LOGNAME={{ .pusher.name }}" ]
AllowEvent = ["push"]
Timeout = 25

Commands = [
  ["git", "pull"],
  ["rsync", "-r", ".", "{{ .repository.name }}.company.com:/opt/var/files"],
]

[[Hook]]
# Here we have some script that can just process the JSON.
Url = "/record-issue"
Commands = [ [ "processevent", "{{ json . }}" ] ]

```

## Acknowledgements

* TOML parser from [BurntSushi/toml](https://github.com/BurntSushi/toml)
* Logging code from my [slightly modified fork](https://github.com/dimfeld/glog) of [Zenoss's heavily modified glog](https://github.com/zenoss/glog)

This project also uses my [httpmuxtree](https://github.com/dimfeld/httptreemux) and [goconfig](https://github.com/dimfeld/goconfig) libraries.
