<div align="center">

<picture>
  <source srcset="img/rect/ahoy-logo-rect-v2.webp" type="image/webp">
  <img src="img/rect/ahoy-logo-rect-v2.svg" alt="Ahoy logo" width="300">
</picture>

<h1>Ahoy!</h1>

<h3>Automate and organise your workflows, no matter what technology you use.</h3>

[![Build and test](https://github.com/ahoy-cli/ahoy/actions/workflows/build_and_test.yml/badge.svg)](https://github.com/ahoy-cli/ahoy/actions/workflows/build_and_test.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/ahoy-cli/ahoy)](https://goreportcard.com/report/github.com/ahoy-cli/ahoy)
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-14-orange.svg)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

</div>

Ahoy is a command line tool that gives each of your projects its own CLI app with zero code and dependencies.

Write your commands in a YAML file and then Ahoy gives you lots of features like:
- a command listing
- per-command help text
- command tab completion
- run commands from any subdirectory

Ahoy makes it easy to create aliases and templates for commands that are useful. It was created to help with running interactive commands within Docker containers, but it's just as useful for local commands, commands over `ssh`, or really anything that could be run from the command line in a single clean interface.

## What's New in v3

Ahoy v3 is a major internal rewrite that brings improved CLI handling whilst maintaining **full backwards compatibility** with all existing `.ahoy.yml` configuration files. Your workflows will not break.

### Key Changes

- **New CLI framework** - Migrated from `urfave/cli` to [Cobra](https://github.com/spf13/cobra), providing a more robust and maintainable foundation.
- **`ahoy config` subcommand group** - Built-in management commands are now grouped under `ahoy config`:
  - `ahoy config init [url]` - Download an example config to get started (replaces the old `ahoy init`).
  - `ahoy config validate` - Check your `.ahoy.yml` for issues and get actionable suggestions.
- **Command descriptions** - Commands now support a separate `description` field for longer, multiline help text, in addition to the existing `usage` field for short summaries.
- **Optional imports** - Import commands can now be marked with `optional: true` so that missing import files are gracefully skipped instead of causing errors.
- **Command aliases** - Commands support an `aliases` field for alternative names, displayed inline in help output.
- **Multiple environment files** - The `env` field now accepts an array of files at both global and command level.
- **Runtime environment variables** - Ahoy injects `AHOY_COMMAND_NAME` (the command being run) and `AHOY_CMD` (path to the ahoy binary) into every command's environment.
- **Improved help output** - Custom help template displays command aliases inline for better discoverability.
- **Full backwards compatibility** - Existing `.ahoy.yml` files continue to work without modification. The YAML API version remains `v2`.

### Upgrading from v2

No changes to your `.ahoy.yml` files are required. Simply replace the `ahoy` binary with the v3 version. All existing commands, aliases, imports, entrypoints, and environment file configurations will continue to work as before.

The only behavioural change you may notice is that `ahoy init` now prints a deprecation notice and redirects to `ahoy config init`. Both work identically.

## Examples

Say you want to import a MySQL database running in `docker-compose` using another container called `cli`. The command could look like this:

`docker exec -i $(docker-compose ps -q cli) bash -c 'mysql -u$DB_ENV_MYSQL_USER -p$DB_ENV_MYSQL_PASSWORD -h$DB_PORT_3306_TCP_ADDR $DB_ENV_MYSQL_DATABASE' < some-database.sql`

With Ahoy, you can turn this into:

`ahoy mysql-import < some-database.sql`

## Quick Start

Get started immediately with our comprehensive examples file:

```bash
# Create a new project with example commands
ahoy config init

# Or download the examples file directly
curl -o .ahoy.yml https://raw.githubusercontent.com/ahoy-cli/ahoy/master/examples/examples.ahoy.yml
```

The examples file includes **30+ usable example commands** for:
- **Local Development Environments** - `up`, `down`, `restart`, `status`
- **Testing & Quality** - `test`, `lint` with multi-language support
- **Database Operations** - `db`, `db:backup` for MySQL/PostgreSQL
- **Build & Deployment** - `build`, `deploy` with safety checks
- **Drupal Integration** - `drush`, `cr`, `uli`, `cex`, `cim` for Drupal projects

**[View the complete examples file](examples/examples.ahoy.yml)**

Try it out:
```bash
ahoy status    # Show service status
ahoy urls      # Show available URLs
ahoy shell     # Open a shell in your container
```

## Features

- Non-invasive - Use your existing workflow! It can wrap commands and scripts you are already using.
- Consistent - Commands always run relative to the `.ahoy.yml` file, but can be called from any subfolder.
- Visual - See a list of all your commands in one place, along with helpful descriptions.
- Flexible - Commands are specific to a single folder tree, so each repo/workspace can have its own commands.
- Command templates - Use regular `bash` syntax like `"$@"` for all arguments, or `$1` for the first argument.
- Fully interactive - Your shells (like MySQL) and prompts still work.
- Import multiple config files using the `imports` field.
- Uses the "last in wins" rule to deal with duplicate commands amongst config files.
- [Command aliases](#command-aliases) - oft-used or long commands can have aliases.
- [Command descriptions](#command-descriptions) - commands can have both short usage text and longer multiline descriptions.
- [Optional imports](#optional-imports) - import commands can gracefully handle missing files.
- [Config validation](#config-validation) - `ahoy config validate` checks your config and reports issues.
- Use a different entrypoint (the thing that runs your commands) if you wish, instead of `bash`. E.g. using PHP, Node.js, Python, etc.
- Plugins are possible by overriding the entrypoint.
- Self-documenting - Commands and help declared in `.ahoy.yml` show up as ahoy command help and [shell completion](#shell-autocompletions) of commands is also available. We have a dedicated Zsh plugin for completions at [ahoy-cli/zsh-ahoy](https://github.com/ahoy-cli/zsh-ahoy).
- [Environment variables](#environment-variables) at both file and command level using the `env` field, with support for multiple env files.
- Runtime variables - `AHOY_COMMAND_NAME` and `AHOY_CMD` are injected into every command so scripts can introspect how they were invoked.

## Installation

### macOS

Using Homebrew / Linuxbrew:

```
brew install ahoy
```

### Linux

Download the [latest release from GitHub](https://github.com/ahoy-cli/ahoy/releases), move the appropriate binary for your platform into someplace in your $PATH and rename it `ahoy`.

Example:
```
os=$(uname -s | tr '[:upper:]' '[:lower:]') && architecture=$(case $(uname -m) in x86_64 | amd64) echo "amd64" ;; aarch64 | arm64 | armv8) echo "arm64" ;; *) echo "amd64" ;; esac) && sudo wget -q https://github.com/ahoy-cli/ahoy/releases/latest/download/ahoy-bin-$os-$architecture -O /usr/local/bin/ahoy && sudo chown $USER /usr/local/bin/ahoy && chmod +x /usr/local/bin/ahoy
```

### Windows

For WSL2, use the Linux binary above for your architecture.

## Command Descriptions

Commands support both a short `usage` field and a longer `description` field. The `usage` appears in the command listing, whilst the `description` provides detailed help text when viewing a specific command.

```yaml
ahoyapi: v2
commands:
  deploy:
    usage: Deploy the application
    description: |
      Deploys the application to the configured environment.

      This command will:
      - Build the production assets
      - Run database migrations
      - Clear all caches
      - Notify the deployment channel

      Use with caution in production environments.
    cmd: ./scripts/deploy.sh
```

## Environment Variables

Ahoy supports loading environment variables from files at both global and command levels, with support for multiple environment files.

#### Single Environment File (backwards compatible):

```yaml
ahoyapi: v2

# Global environment file relative to .ahoy.yml
env: .env

commands:
  db-import:
    # Command-specific environment file, overrides global vars
    env: .env.db
    usage: Import a database
    cmd: mysql -u$DB_USER -p$DB_PASSWORD $DB_NAME < $1
```

#### Multiple Environment Files:

```yaml
ahoyapi: v2

# Multiple global environment files loaded in order
env:
  - .env.base
  - .env.local
  - .env.override

commands:
  deploy:
    # Multiple command-specific env files
    env:
      - .env.deploy
      - .env.secrets
    usage: Deploy the application
    cmd: ./deploy.sh
```

#### Environment File Format:
```sh
# Global .env file
DB_USER=root
DB_PASSWORD=root

# Command-specific .env.db file
DB_USER=custom_user
DB_PASSWORD=secret
DB_NAME=mydb
```

**Key Features:**
- Files are loaded in order, with later files overriding earlier ones.
- Command-level env files override global env files.
- Non-existent files are gracefully ignored.
- Supports comments and empty lines in env files.
- Maintains full backwards compatibility with single file syntax.

#### Runtime Environment Variables

Ahoy automatically injects two variables into every command's environment:

| Variable | Value |
|---|---|
| `AHOY_COMMAND_NAME` | The name of the command being run |
| `AHOY_CMD` | Path to the ahoy binary |

These are useful for scripts that need to know how they were invoked, or that want to call other ahoy commands via `$AHOY_CMD`.

## Command Aliases

Ahoy supports command aliases, allowing you to define alternative names for your commands.

### Usage

In your `.ahoy.yml` file, add an `aliases` field to any command definition:

```yaml
ahoyapi: v2
commands:
  hello:
    usage: Say hello
    cmd: echo "Hello, World!"
    aliases: ["hi", "greet"]
```

In this example, the `hello` command can also be invoked using `hi` or `greet`.

### Notes

- Aliases are displayed in the help output next to each command.
- Bash completion works with aliases as well as primary command names.
- **If multiple commands share the same alias, the "last in wins" rule is used.**

## Optional Imports

Import commands can be marked as optional, allowing missing import files to be gracefully skipped rather than causing a fatal error. This is useful for separating commands into public and private sets, or for supporting optional tooling.

```yaml
ahoyapi: v2
commands:
  local-tools:
    usage: Local development tools
    optional: true
    imports:
      - ./local-tools.ahoy.yml
      - ./team-tools.ahoy.yml

  core-tools:
    usage: Core project tools
    imports:
      - ./core.ahoy.yml
```

If `optional: true` is set and none of the imported files can be found, the command is silently omitted from the command listing. Without `optional`, missing imports will produce a fatal error.

## Config Validation

Ahoy v3 includes a built-in configuration validator:

```bash
ahoy config validate
```

This checks your `.ahoy.yml` (and any imported files) for common issues, including:

- Unsupported fields or YAML API version mismatches
- Missing import files (without `optional: true`)
- Features that require a newer version of Ahoy

Validation warnings are shown in verbose mode (`-v`); errors are always shown. The validator also provides actionable suggestions when it finds a problem.

## Shell Autocompletions

### Zsh

For Zsh completions, we have a standalone plugin available at [ahoy-cli/zsh-ahoy](https://github.com/ahoy-cli/zsh-ahoy).

### Bash

For Bash, you'll need to make sure you have bash-completion installed and set up. See [bash/zsh completion](https://ahoy-cli.readthedocs.io/en/latest/#bash-zsh-completion) for further instructions.

## Example of the YAML File Setup

```YAML
# All files must have v2 set or you'll get an error.
ahoyapi: v2

# You can override the entrypoint. This is the default if you don't override it.
# {{cmd}} is replaced with your command and {{name}} is the name of the command that was run (available as $0).
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
      description: |
        Demonstrates multi-line commands, parameter passing,
        and calling other ahoy commands from within a command.
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

## Sponsors

- [<img src="https://raw.githubusercontent.com/drevops/website/refs/heads/develop/web/themes/custom/drevops/assets/logos/logo_primary_light_desktop.svg?sanitize=true" width="160px;" alt="DrevOps Logo"><br />Alex Skrypnyk - DrevOps](https://drevops.com)

## Contributors

Thanks to all these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tbody>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/acouch"><img src="https://avatars.githubusercontent.com/u/512243?v=4?s=64" width="64px;" alt="Aaron Couch"/><br /><sub><b>Aaron Couch</b></sub></a><br /><a href="https://github.com/ahoy-cli/Ahoy/commits?author=acouch" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/aashil"><img src="https://avatars.githubusercontent.com/u/6216576?v=4?s=64" width="64px;" alt="Aashil Patel"/><br /><sub><b>Aashil Patel</b></sub></a><br /><a href="https://github.com/ahoy-cli/Ahoy/commits?author=aashil" title="Code">💻</a> <a href="https://github.com/ahoy-cli/Ahoy/commits?author=aashil" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://www.drevops.com/"><img src="https://avatars.githubusercontent.com/u/378794?v=4?s=64" width="64px;" alt="Alex Skrypnyk"/><br /><sub><b>Alex Skrypnyk</b></sub></a><br /><a href="https://github.com/ahoy-cli/Ahoy/issues?q=author%3AAlexSkrypnyk" title="Bug reports">🐛</a> <a href="https://github.com/ahoy-cli/Ahoy/pulls?q=is%3Apr+reviewed-by%3AAlexSkrypnyk" title="Reviewed Pull Requests">👀</a> <a href="#question-AlexSkrypnyk" title="Answering Questions">💬</a> <a href="#promotion-AlexSkrypnyk" title="Promotion">📣</a> <a href="#ideas-AlexSkrypnyk" title="Ideas, Planning, & Feedback">🤔</a> <a href="#financial-AlexSkrypnyk" title="Financial">💵</a> <a href="#security-AlexSkrypnyk" title="Security">🛡️</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://www.linkedin.com/in/alexandrerafalovitch"><img src="https://avatars.githubusercontent.com/u/64153?v=4?s=64" width="64px;" alt="Alexandre Rafalovitch"/><br /><sub><b>Alexandre Rafalovitch</b></sub></a><br /><a href="https://github.com/ahoy-cli/Ahoy/commits?author=arafalov" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/hanoii"><img src="https://avatars.githubusercontent.com/u/677879?v=4?s=64" width="64px;" alt="Ariel Barreiro"/><br /><sub><b>Ariel Barreiro</b></sub></a><br /><a href="https://github.com/ahoy-cli/Ahoy/commits?author=hanoii" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://agaric.coop/"><img src="https://avatars.githubusercontent.com/u/27131?v=4?s=64" width="64px;" alt="Benjamin Melançon"/><br /><sub><b>Benjamin Melançon</b></sub></a><br /><a href="https://github.com/ahoy-cli/Ahoy/commits?author=mlncn" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/ocean"><img src="https://avatars.githubusercontent.com/u/4443?v=4?s=64" width="64px;" alt="Drew Robinson"/><br /><sub><b>Drew Robinson</b></sub></a><br /><a href="https://github.com/ahoy-cli/Ahoy/commits?author=ocean" title="Code">💻</a> <a href="https://github.com/ahoy-cli/Ahoy/issues?q=author%3Aocean" title="Bug reports">🐛</a> <a href="#content-ocean" title="Content">🖋</a> <a href="https://github.com/ahoy-cli/Ahoy/commits?author=ocean" title="Documentation">📖</a> <a href="#ideas-ocean" title="Ideas, Planning, & Feedback">🤔</a> <a href="#infra-ocean" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a> <a href="#maintenance-ocean" title="Maintenance">🚧</a> <a href="#platform-ocean" title="Packaging/porting to new platform">📦</a> <a href="#question-ocean" title="Answering Questions">💬</a> <a href="https://github.com/ahoy-cli/Ahoy/pulls?q=is%3Apr+reviewed-by%3Aocean" title="Reviewed Pull Requests">👀</a> <a href="#security-ocean" title="Security">🛡️</a> <a href="https://github.com/ahoy-cli/Ahoy/commits?author=ocean" title="Tests">⚠️</a></td>
    </tr>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://www.elijahlynn.net/"><img src="https://avatars.githubusercontent.com/u/1504756?v=4?s=64" width="64px;" alt="Elijah Lynn"/><br /><sub><b>Elijah Lynn</b></sub></a><br /><a href="https://github.com/ahoy-cli/Ahoy/commits?author=ElijahLynn" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://botsandbrains.com/"><img src="https://avatars.githubusercontent.com/u/377330?v=4?s=64" width="64px;" alt="Frank Carey"/><br /><sub><b>Frank Carey</b></sub></a><br /><a href="https://github.com/ahoy-cli/Ahoy/commits?author=frankcarey" title="Code">💻</a> <a href="https://github.com/ahoy-cli/Ahoy/issues?q=author%3Afrankcarey" title="Bug reports">🐛</a> <a href="#content-frankcarey" title="Content">🖋</a> <a href="https://github.com/ahoy-cli/Ahoy/commits?author=frankcarey" title="Documentation">📖</a> <a href="#ideas-frankcarey" title="Ideas, Planning, & Feedback">🤔</a> <a href="#infra-frankcarey" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a> <a href="#maintenance-frankcarey" title="Maintenance">🚧</a> <a href="#platform-frankcarey" title="Packaging/porting to new platform">📦</a> <a href="#question-frankcarey" title="Answering Questions">💬</a> <a href="https://github.com/ahoy-cli/Ahoy/pulls?q=is%3Apr+reviewed-by%3Afrankcarey" title="Reviewed Pull Requests">👀</a> <a href="#security-frankcarey" title="Security">🛡️</a> <a href="https://github.com/ahoy-cli/Ahoy/commits?author=frankcarey" title="Tests">⚠️</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/jackwrfuller"><img src="https://avatars.githubusercontent.com/u/78133717?v=4?s=64" width="64px;" alt="Jack Fuller"/><br /><sub><b>Jack Fuller</b></sub></a><br /><a href="https://github.com/ahoy-cli/Ahoy/issues?q=author%3Ajackwrfuller" title="Bug reports">🐛</a> <a href="https://github.com/ahoy-cli/Ahoy/commits?author=jackwrfuller" title="Code">💻</a> <a href="https://github.com/ahoy-cli/Ahoy/commits?author=jackwrfuller" title="Documentation">📖</a> <a href="https://github.com/ahoy-cli/Ahoy/commits?author=jackwrfuller" title="Tests">⚠️</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/jnsalsa"><img src="https://avatars.githubusercontent.com/u/194740356?v=4?s=64" width="64px;" alt="Jonathan Nagy"/><br /><sub><b>Jonathan Nagy</b></sub></a><br /><a href="https://github.com/ahoy-cli/Ahoy/issues?q=author%3Ajnsalsa" title="Bug reports">🐛</a> <a href="https://github.com/ahoy-cli/Ahoy/commits?author=jnsalsa" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://msound.net/"><img src="https://avatars.githubusercontent.com/u/432912?v=4?s=64" width="64px;" alt="Mani Soundararajan"/><br /><sub><b>Mani Soundararajan</b></sub></a><br /><a href="https://github.com/ahoy-cli/Ahoy/commits?author=msound" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://morpht.com/"><img src="https://avatars.githubusercontent.com/u/1254919?v=4?s=64" width="64px;" alt="Marji Cermak"/><br /><sub><b>Marji Cermak</b></sub></a><br /><a href="https://github.com/ahoy-cli/Ahoy/commits?author=marji" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/dkinzer"><img src="https://avatars.githubusercontent.com/u/444215?v=4?s=64" width="64px;" alt="david kinzer (he/him)"/><br /><sub><b>david kinzer (he/him)</b></sub></a><br /><a href="https://github.com/ahoy-cli/Ahoy/commits?author=dkinzer" title="Code">💻</a></td>
    </tr>
  </tbody>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!
