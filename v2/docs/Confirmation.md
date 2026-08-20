Confirmation allows you to specify a string to prompt the user with and then control the execution of your command. If they type "y", the command will return true (exit 0), otherwise it will return false (exit 1).

```Yaml
...
  commands:
    confirm:
      cmd: |
        read -r -p "{{args}} [y/N] " response
        if [ $response = y ]
        then
          true
        else
          false
        fi
```

### Examples

Simple confirmation
```
$ ahoy confirm "Are you sure you want to do this?"
Are you sure you want to do this? [y/N]
$ y
```

Using confirmation with other commands. See [[Controlling Execution]]
```
$ ahoy confirm "Delete your /tmp directory?" && \
  rm -rf /tmp/* || \
  echo "Skipping..."
  Delete your /tmp directory? [y/N]
$ n # (or anything besides y)
  Skipping...
```

