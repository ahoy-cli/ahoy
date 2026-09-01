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

**[📖 Read the full documentation at ahoy-cli.github.io](https://ahoy-cli.github.io/)**

</div>

Ahoy gives each of your projects its own CLI app with zero code and zero dependencies. Write your commands in a YAML file, and Ahoy turns them into a real command line tool with a command listing, per-command help, shell tab completion, and the ability to run your commands from any subdirectory.

It was created to help with running interactive commands inside Docker containers, but it works just as well for local commands, aliases for commands with complex parameters, commands over `ssh`, or anything else you would otherwise run by hand.

## Why

Say you want to import a MySQL database running in `docker-compose` via a container called `cli`. Without Ahoy:

```bash
docker exec -i $(docker-compose ps -q cli) bash -c 'mysql -u$DB_ENV_MYSQL_USER -p$DB_ENV_MYSQL_PASSWORD -h$DB_PORT_3306_TCP_ADDR $DB_ENV_MYSQL_DATABASE' < some-database.sql
```

With Ahoy:

```bash
ahoy mysql-import < some-database.sql
```

## Installation

**macOS** - using Homebrew / Linuxbrew:

```bash
brew install ahoy
```

**Linux** - download the [latest release](https://github.com/ahoy-cli/ahoy/releases), put the binary for your platform somewhere in your `$PATH` and rename it `ahoy`. Or use the one-liner:

```bash
os=$(uname -s | tr '[:upper:]' '[:lower:]') && architecture=$(case $(uname -m) in (x86_64 | amd64) echo "amd64" ;; (aarch64 | arm64 | armv8*) echo "arm64" ;; (armv7*) echo "armv7" ;; (armv6*) echo "armv6" ;; esac) && { [ -n "$architecture" ] || { echo "Unsupported architecture: $(uname -m)" >&2; false; }; } && sudo wget -q https://github.com/ahoy-cli/ahoy/releases/latest/download/ahoy-bin-$os-$architecture -O /usr/local/bin/ahoy && sudo chown $USER /usr/local/bin/ahoy && chmod +x /usr/local/bin/ahoy
```

**Windows** - native builds are published as `ahoy-bin-windows-amd64.exe` and `ahoy-bin-windows-arm64.exe` on the [latest release](https://github.com/ahoy-cli/ahoy/releases) page (also available as `ahoy-windows-amd64.zip` / `ahoy-windows-arm64.zip`). Download the one for your architecture, rename it `ahoy.exe`, and put it in a directory that is on your `PATH` - for example `%USERPROFILE%\bin`, adding that folder under Settings → System → About → Advanced system settings → Environment Variables. Note that Ahoy runs command bodies through `bash -c` by default, so bash needs to be available (Git Bash, for instance) unless your config sets a custom `entrypoint`.

For WSL2, use the Linux binary above.

Full instructions: **[Installation & Setup](https://ahoy-cli.github.io/guides/getting-started/)**

## Quick start

```bash
# Download an example config into your project
ahoy config init

# See what you can run
ahoy
```

The example file ships with more than 20 ready-to-use commands covering local environments (`up`, `down`, `restart`), testing and linting, database operations, build and deploy, and Drupal integration. **[View it here](examples/examples.ahoy.yml)**.

## What a config looks like

```yaml
# All files must have v2 set or you'll get an error.
ahoyapi: v2

# Optional: load environment variables. Accepts one file or a list.
env: .env

commands:
  simple-command:
    usage: An example of a single-line command.
    cmd: echo "Do stuff with bash"

  deploy:
    usage: Deploy the application
    aliases: ["dep"]
    description: |
      Deploys the application to the configured environment.
      Builds assets, runs migrations, and clears caches.
    cmd: ./scripts/deploy.sh

  multi-line:
    usage: Show more advanced features.
    cmd: |
      echo "multi-line bash script";
      ahoy simple-command          # call other ahoy commands
      echo "your params were: $@"  # standard bash argument handling
      echo "param1: $1"

  subcommands:
    usage: Group commands from other config files.
    # Later files override earlier ones. Add `optional: true` to tolerate
    # missing files.
    imports:
      - ./some-file1.ahoy.yml
      - ./some-file2.ahoy.yml
```

Every field is documented in the **[YAML Schema reference](https://ahoy-cli.github.io/reference/yaml-schema/)**.

## Features

- **Non-invasive** - wraps the commands and scripts you already use.
- **Consistent** - commands always run relative to the `.ahoy.yml` file, but can be called from any subfolder.
- **Visual** - see all your commands in one place with helpful descriptions.
- **Flexible** - each repo or workspace gets its own commands.
- **Fully interactive** - shells like MySQL and interactive prompts still work.
- **Command templates** - use regular bash syntax like `"$@"` or `$1`.
- **[Imports](https://ahoy-cli.github.io/guides/importing/)** - split commands across multiple files, with "last in wins" for duplicates, and `optional: true` to skip missing files gracefully.
- **[Aliases and descriptions](https://ahoy-cli.github.io/guides/writing-commands/)** - give commands short alternative names and longer multi-line help text.
- **[Environment variables](https://ahoy-cli.github.io/guides/environment/)** - load one or more env files at both file and command level. Ahoy also injects `AHOY_COMMAND_NAME` and `AHOY_CMD` into every command.
- **[Shell completion](https://ahoy-cli.github.io/guides/shell-autocompletion/)** - commands and help are self-documenting. There's a dedicated Zsh plugin at [ahoy-cli/zsh-ahoy](https://github.com/ahoy-cli/zsh-ahoy).
- **Custom entrypoints** - swap `bash` for PHP, Node.js, Python or anything else. This is also how plugins work.
- **Config validation** - `ahoy config validate` checks your config and suggests fixes.

## What's new in v3

Ahoy v3 is a major internal rewrite that brings improved CLI handling whilst maintaining **full backwards compatibility** with existing `.ahoy.yml` files. Your workflows will not break.

- **New CLI framework** - migrated from `urfave/cli` to [Cobra](https://github.com/spf13/cobra) for a more robust foundation.
- **`ahoy config` subcommand group** - `ahoy config init [url]` downloads an example config (replacing `ahoy init`), and `ahoy config validate` checks your config for issues.
- **Command descriptions** - a `description` field for longer multi-line help, alongside the existing short `usage` field.
- **Optional imports** - mark imports with `optional: true` so missing files are skipped instead of erroring.
- **Command aliases** - an `aliases` field for alternative names, shown inline in help output.
- **Multiple environment files** - `env` now accepts an array, at both global and command level.
- **Runtime environment variables** - `AHOY_COMMAND_NAME` and `AHOY_CMD` are injected into every command.

### Upgrading from v2

No changes to your `.ahoy.yml` files are required - just replace the binary. All existing commands, aliases, imports, entrypoints and environment configuration continue to work, and the YAML API version stays at `v2`.

The one behavioural change: `ahoy init` and `ahoy config init` create exactly the same configuration file, but `ahoy init` now also prints a deprecation notice pointing you at `ahoy config init`.

## Documentation

Full documentation lives at **[ahoy-cli.github.io](https://ahoy-cli.github.io/)**:

| | |
|---|---|
| [Installation & Setup](https://ahoy-cli.github.io/guides/getting-started/) | Getting Ahoy onto your machine |
| [Writing Commands](https://ahoy-cli.github.io/guides/writing-commands/) | Usage text, descriptions, aliases, arguments |
| [Command Execution](https://ahoy-cli.github.io/guides/command-execution/) | How commands run, chaining, entrypoints |
| [Importing & Overriding](https://ahoy-cli.github.io/guides/importing/) | Splitting configs across files |
| [Environment](https://ahoy-cli.github.io/guides/environment/) | Env files and runtime variables |
| [Shell Autocompletion](https://ahoy-cli.github.io/guides/shell-autocompletion/) | Bash and Zsh completions |
| [CLI Reference](https://ahoy-cli.github.io/reference/cli/) | Every command and flag |
| [YAML Schema](https://ahoy-cli.github.io/reference/yaml-schema/) | Every configuration field |

## Planned features

- Specify arguments and flags in the ahoy file itself, to cut down on argument parsing in scripts.
- A "verify" YAML option creating a yes/no prompt for potentially destructive commands.
- Pipe tab completion to another command.

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
