This actually works as-is using the yaml `|` pipe syntax..



```Yaml
...
  commands:
    my-multiline-command:
      cmd: |
        echo "1 - If this passes" &&
        echo "2 - Then do this" ||
        echo "3 - Or do this if 1 or 2 fails (returns non-zero)"
```

You can even write full scripts:

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

If you want to do multiple lines that append together (a really long string perhaps), you can use the Yaml `-` dash character.

```Yaml
...
  commands:
    echo-long-string:
      cmd: -
        echo "This is a really really
        really really long string"
```