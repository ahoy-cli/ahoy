![AHOY logo](https://avatars.githubusercontent.com/u/19353604?s=250&v=4)

# AHOY! - Automate and organize your workflows, no matter what technology you use.

[![CircleCI](https://dl.circleci.com/status-badge/img/gh/ahoy-cli/ahoy/tree/master.svg?style=shield)](https://dl.circleci.com/status-badge/redirect/gh/ahoy-cli/ahoy/tree/master) [![Go Report Card](https://goreportcard.com/badge/github.com/ahoy-cli/ahoy)](https://goreportcard.com/report/github.com/ahoy-cli/ahoy)

### Note: Ahoy 2.x is now released and is the only supported version.

Ahoy is command line tool that gives each of your projects their own CLI app with zero code and dependencies.

Write your commands in a YAML file and then Ahoy gives you lots of features like:
* a command listing
* per-command help text
* command tab completion
* run commands from any subdirectory

Ahoy makes it easy to create aliases and templates for commands that are useful. It was created to help with running interactive commands within Docker containers, but it's just as useful for local commands, commands over `ssh`, or really anything that could be run from the command line in a single clean interface.

## Examples

Say you want to import a MySQL database running in `docker-compose` using another container called `cli`. The command could look like this:

`docker exec -i $(docker-compose ps -q cli) bash -c 'mysql -u$DB_ENV_MYSQL_USER -p$DB_ENV_MYSQL_PASSWORD -h$DB_PORT_3306_TCP_ADDR $DB_ENV_MYSQL_DATABASE' < some-database.sql`

With Ahoy, you can turn this into:

`ahoy mysql-import < some-database.sql`

[More examples](https://ahoy-cli.readthedocs.io/en/latest/Home.html).

## Features

- Non-invasive - Use your existing workflow! It can wrap commands and scripts you are already using.
- Consistent - Commands always run relative to the `.ahoy.yml` file, but can be called from any subfolder.
- Visual - See a list of all your commands in one place, along with helpful descriptions.
- Flexible - Commands are specific to a single folder tree, so each repo/workspace can have its own commands.
- Command templates - Args can be dropped into your commands using `{{args}}`
- Fully interactive - Your shells (like MySQL) and prompts still work.
- Self-documenting - Commands and help declared in `.ahoy.yml` show up as ahoy command help and shell completion of commands (see [bash/zsh completion](https://ahoy-cli.readthedocs.io/en/latest/#bash-zsh-completion)) is also available.

## Installation

### macOS

Using Homebrew:

```
brew install ahoy
```

Note that `ahoy` is in `homebrew-core` as of 1/18/19, so you don't need to use the old tap.
If you were previously using it, you can use the following command to remove it:

```
brew untap ahoy-cli/tap
```

### Linux

Download the [latest release from GitHub](https://github.com/ahoy-cli/ahoy/releases), move the appropriate binary for your plaform into someplace in your $PATH and rename it `ahoy`.

Example:
```
os=$(uname -s | tr [:upper:] [:lower:]) && architecture=$(case $(uname -m) in x86_64 | amd64) echo "amd64" ;; aarch64 | arm64 | armv8) echo "arm64" ;; *) echo "amd64" ;; esac) && sudo wget -q https://github.com/ahoy-cli/ahoy/releases/latest/download/ahoy-bin-$os-$architecture -O /usr/local/bin/ahoy && sudo chown $USER /usr/local/bin/ahoy && chmod +x /usr/local/bin/ahoy
```

### Windows

For WSL2, use the Linux binary above for your architecture.

## Command Aliases

Ahoy now supports command aliases. This feature allows you to define alternative names for your commands, making them more convenient to use and remember.

### Usage

In your `.ahoy.yml` file, you can add an `aliases` field to any command definition. The `aliases` field should be an array of strings, each representing one or more alternative names for the command.

Example:

```yaml
ahoyapi: v2
commands:
  hello:
    usage: Say hello
    cmd: echo "Hello, World!"
    aliases: ["hi", "greet"]
```

In this example, the `hello` command can also be invoked using `hi` or `greet`.

### Benefits

- Improved usability: Users can call commands using shorter or more intuitive names.
- Flexibility: You can provide multiple ways to access the same functionality without duplicating command definitions.
- Backward compatibility: You can introduce new, more descriptive command names while keeping old names as aliases.

### Notes

- Aliases are displayed in the help output for each command.
- Bash completion works with aliases as well as primary command names.
- **If multiple commands share the same alias, the "last in wins" rule is used and the last matching command will be executed.**

## Some additions in v2

- Implements a new feature to import multiple config files using the "imports" field.
- Uses the "last in wins" rule to deal with duplicate commands amongst the config files.
- Better handling of quotes by no longer using `{{args}}`. Use regular `bash` syntax like `"$@"` for all arguments, or `$1` for the first argument.
- You can now use a different entrypoint (the thing that runs your commands) instead of `bash`. E.g. using PHP, Node.js, Python, etc. is possible.
- Plugins are now possible by overriding the entrypoint.

### Example of new YAML setup in v2

```YAML
# All files must have v2 set or you'll get an error
ahoyapi: v2

# You can now override the entrypoint. This is the default if you don't override it.
# {{cmd}} is replaced with your command and {{name}} is the name of the command that was run (available as $0)
entrypoint:
  - bash
  - "-c"
  - '{{cmd}}'
  - '{{name}}'
commands:
  simple-command:
      usage: An example of a single-line command.
      cmd: echo "Do stuff with bash"

  complex-command:
      usage: Show more advanced features.
      cmd: | # We support multi-line commands with pipes.
          echo "multi-line bash script";
          # You can call other ahoy commands.
          ahoy simple-command
          # you can take params
          echo "your params were: $@"
          # you can use numbered params, same as bash.
          echo "param1: $1"
          echo "param2: $2"
          # Everything bash supports is available, if statements, etc.
          # Hate bash? Use something else like python in a subscript or change the entrypoint.

  subcommands:
      usage: List the commands from the imported config files.
      # These commands will be aggregated together with later files overriding earlier ones if they exist.
      imports:
        - ./some-file1.ahoy.yml
        - ./some-file2.ahoy.yml
        - ./some-file3.ahoy.yml
```

## Planned Features

- Enable specifying specific arguments and flags in the ahoy file itself to cut down on parsing arguments in scripts.
- Support for more built-in commands or a "verify" YAML option that would create a yes / no prompt for potentially destructive commands. (Are you sure you want to delete all your containers?)
- Pipe tab completion to another command (allows you to get tab completion).
- Support for configuration.
