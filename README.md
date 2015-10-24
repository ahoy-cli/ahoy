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

## TODOS

- Disable logging unless verbose mode is set.
- Provide "drivers" for bash, docker-compose, kubernetes
- Do specific arg replacement like {{arg1}}
- Add ability to use subcommands (would help clean up a longer list of commands)
- Add ability to specify the .ahoy.yml file you want to use, using a flag.
- Support a .ahoy.yml file in your home directory for general use aliases.
- Support multiple line commands (already supported with && but this might be a nice and cleaner option)
