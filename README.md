# AHOY!
ahoy makes is easy to create aliases and templates for commands that are useful. It was specifically created to help with running interactive commands within docker containers, but you can also use it for local commands, commands over ssh, etc.

## Examples

Say you want to import a sql database running in docker-compose using another container called cli. The command could look like this:
`docker exec -i $(docker-compose ps -q cli) bash -c 'mysql -u$DB_ENV_MYSQL_USER -p$DB_ENV_MYSQL_PASSWORD -h$DB_PORT_3306_TCP_ADDR $DB_ENV_MYSQL_DATABASE' < some-database.sql`
With ahoy, you can turn this into
`ahoy mysql-import < some-database.sql`

## FEATURES
- Easily create shortcuts in a yml file that run relative to the .ahoy.yml file.
- Visualize a list of all of your alias commands in one place, along with helpful descriptions.
- Aliases are specific to a single folder tree, so each repo/workspace can have its own commands
- Arg replacement in commands using {{args}}
- Fully interactive shells
- Commands and help declared in .ahoy.yml show up as ahoy command help
- ahoy can be run in any subfolder where the yaml file exists

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
- Provide "drivers" for bash, docker-compose, kubernetes
- Do specific arg replacement like {{arg1}}
- Add ability to use subcommands (would help clean up a longer list of commands)
- Add ability to specify the .ahoy.yml file you want to use, using a flag.
- Support a .ahoy.yml file in your home directory for general use aliases.
- Support multiple line commands (already supported with && but this might be a nice and cleaner option)
- Upload individual binaries instead of a tar ball of each one. (4.6M each, vs 2M for the tarball)
- Support alternate init files (take a url parameter) which give a way to deploy the config files.
