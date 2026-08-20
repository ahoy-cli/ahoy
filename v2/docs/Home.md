Welcome to the ahoy wiki!


## Ahoy command snippets examples
* [Basic Tips and Troubleshooting](Basic-Tips-and-Troubleshooting.html)
* [Confirmation](Confirmation.html)
* [Controlling Execution](Controlling-Execution.html)
* [Multiple line commands](Multiple-line-commands.html)

## Basics
The simplest way to start using ahoy is to create a basic .ahoy.yml file in your current directory like so:
```Yaml
ahoyapi: v1
usage: DKAN cli app for development using ahoy.
commands:
  echo:
    usage: Simply echo all the arguments
    # Note that {{args}} will be replaced with the string of all arguments passed
    cmd: echo "{{args}}"
```

Now if we simply run ahoy, it will find that file and output the help text and a list of commands.
```
$ ahoy
NAME:
   ahoy - DKAN cli app for development using ahoy.
USAGE:
   ahoy [global options] command [command options] [arguments...]
COMMANDS:
   echo	Simply echo all the arguments
   init	Initialize a new .ahoy.yml config file in the current directory.
GLOBAL OPTIONS:
   --verbose, -v		Output extra details like the commands to be run. [$AHOY_VERBOSE]
   --file, -f 			Use a specific ahoy file.
   --help, -h			show help
   --version			print the version
   --generate-bash-completion
VERSION:
   0.0.0
```
Now if we call `ahoy -v echo "Do I hear an echo?"` ..
```
ahoy -v echo "Do I hear an echo?"
2016/01/13 14:03:54 ===> Ahoy echo from  : echo "Do I hear an echo?"
Do I hear an echo?
```

### Writing More Complex Commands
Let's show an example of using bash scripts AND reusing ahoy commands
```Yaml
ahoyapi: v1
commands:
  confirm:
    cmd: |
      read -r -p "{{args}} [y/N] " response
      if [ $response = y ]; then
        true
      else
        false
      fi
    # This will keep the confirm command from showing up in the help text.
    hide: true
  meaning-of-life:
    cmd: |
      ahoy confirm "Are you sure you want to know?" &&
      # Run this if confirm returns true
      echo The meaning of life is 42 ||
      # Run this if confirm returns false
      echo "OK, you don't want to know, skipping..."
```
```
$ ahoy meaning-of-life
Are you sure you want to know? [y/N] y
The meaning of life is 42

$ ahoy meaning-of-life
Are you sure you want to know? [y/N] n
OK, you don't want to know, skipping...
```

##Importing commands from other ahoy files.

Another powerful feature is importing commands from other files.
###Subcommands
Ahoy allows you to import an entire yml file full of commands by using `import: relative path to file` instead of `cmd`. This is useful to organize commands into groups.
###Direct import
You can also import single commands by calling ahoy and setting the path of the .ahoy file you want to use using the -f flag.
```Yaml
#sub.ahoy.yml
ahoyapi: v1
commands:
  whoami:
    #Simple unix command that displays the logged in user.
    cmd: whoami
```
```Yaml
#.ahoy.yml
ahoyapi: v1
commands:
  # Imports a single ahoy command called whoami and changes the name to direct example
  direct-example:
    usage: Runs the whoami command directly
    cmd: ahoy -f sub.ahoy.yml whoami
  # Imports all commands in the file
  import-example:
    usage: Loads all commands in a subfile.
    import: sub.ahoy.yml
```
```
$ ahoy direct-example
fcarey #or whatever your user name is

$ ahoy import-example whoami
fcarey #or whatever your user name is

#Shows help text for imported subcommands
$ahoy import-example
NAME:
   ahoy - Creates a configurable cli app for running commands.
USAGE:
   ahoy [global options] command [command options] [arguments...]
COMMANDS:
   whoami
   init		Initialize a new .ahoy.yml config file in the current directory.
GLOBAL OPTIONS:
   --verbose, -v		Output extra details like the commands to be run. [$AHOY_VERBOSE]
   --file, -f 			Use a specific ahoy file.
   --help, -h			show help
   --version			print the version
   --generate-bash-completion
VERSION:
   0.0.0
```
