# AHOY!
ahoy makes is easy to create aliases and templates for commands that are useful. It was specifically created to help with running interactive commands within docker containers, but you can also use it for local commands, commands over ssh, etc.

## Examples

Say you want to import a sql database running in docker-compose using another container called cli. The command could look like this:

`docker exec -i $(docker-compose ps -q cli) bash -c 'mysql -u$DB_ENV_MYSQL_USER -p$DB_ENV_MYSQL_PASSWORD -h$DB_PORT_3306_TCP_ADDR $DB_ENV_MYSQL_DATABASE' < some-database.sql`

With ahoy, you can turn this into

`ahoy mysql-import < some-database.sql`

## FEATURES
- Non-invasive - Use your existing workflow! It can wrap commands and scripts you are already using.
- Consitent - Commands always run relative to the .ahoy.yml file, but can be called from any subfolder.
- Visual - See a list of all of your commands in one place, along with helpful descriptions.
- Flexible - Commands are specific to a single folder tree, so each repo/workspace can have its own commands
- Command Templates - Args can be dropped into your commands using `{{args}}`
- Fully interactive  - your shells (like mysql) and prompts still work.
- Self-Documenting - Commands and help declared in .ahoy.yml show up as ahoy command help and bash completion of commands (see below)

## INSTALLATION

### OSX
Using Homebrew:
```
brew tap devinci-code/tap
brew install ahoy
```

### Linux
Download and unzip the latest release and move the appropriate binary for your plaform into someplace in your $PATH and rename it `ahoy`

Example:
```
wget https://github.com/devinci-code/ahoy/releases/download/1.0.0/ahoy-release-1-0-0.tar.gz
tar xzvf ahoy-release-1-0-0.tar.gz
cp ahoy-release-1-0-0/ahoy-linux-amd64 /usr/local/bin/ahoy
chown +x /usr/local/bin/ahoy
```

### Bash / Zsh Completion
For Zsh, Just add this to your ~/.zshrc, and your completions will be relative to the directory you're in.

`complete -F "ahoy --generate-bash-completion" ahoy`

For Bash, you'll need to make sure you have bash-completion installed and setup. On OSX with homebrew it looks like this:

`brew install bash bash-completion`

Now make sure you follow the couple installation instructions in the "Caveats" section that homebrew returns. And make sure completion is working for git for instance before you continue (you may need to restart your shell)

Then, (for homebrew) you'll want to create a file at `/usr/local/etc/bash_completion.d/ahoy` with the following:

```Bash
_ahoy()
{
    local cur=${COMP_WORDS[COMP_CWORD]}
    COMPREPLY=( $(compgen -W "`ahoy --generate-bash-completion`" -- $cur) )
}
complete -F _ahoy ahoy
```

restart your shell, and you should see ahoy autocomplete when typing `ahoy [TAB]`

## USAGE
Almost all the commands are actually specified in a .ahoy.yml file placed in your working tree somewhere. Commands that are added there show up as options in ahoy. Here is what it looks like when using the [example.ahoy.yml file](https://github.com/devinci-code/ahoy/blob/master/examples/examples.ahoy.yml). To start with this file locally you can run `ahoy init`.

```
$ ahoy
NAME:
   ahoy - Send commands to docker-compose services

USAGE:
   ahoy [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
   vdown	Stop the vagrant box if one exists.
   vup		Start the vagrant box if one exists.
   start	Start the docker compose-containers.
   stop		Stop the docker-compose containers.
   restart	Restart the docker-compose containers.
   drush	Run drush commands in the cli service container.
   bash		Start a shell in the container (like ssh without actual ssh).
   sqlc		Connect to the default mysql database. Supports piping of data into the command.
   behat	Run the behat tests within the container.
   ps		List the running docker-compose containers.
   behat-init	Use composer to install behat dependencies.
   init		Initialize a new .ahoy.yml config file in the current directory.
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h			show help
   --generate-bash-completion
   --version, -v		print the version
```

## TODOS

- Disable logging unless verbose mode is set.
- Provide "drivers" for bash, docker-compose, kubernetes (these systems still work now, this would just make it easier)
- Do specific arg replacement like {{arg1}}
- Add ability to use subcommands (would help clean up a longer list of commands)
- Add ability to specify the .ahoy.yml file you want to use, using a flag.
- Support multiple line commands (already supported with && but this might be a nice and cleaner option)
- Support alternate init files (take a url parameter) which give a way to deploy the config files.
- Add a verbose flag that would actually output the command before it's run.
- Support a "verify" yaml option that would create a yes / no prompt for potentially destructive commands. (Are you sure you want to delete all your containers?)
