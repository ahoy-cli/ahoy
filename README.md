# AHOY!
ahoy makes is easy to create aliases for commands that are useful, but ones you'd rather not type them repeatedly. It was specifically created to help with running interactive commands within docker containers, but you can also use it for local commands, commands over ssh, etc.

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
