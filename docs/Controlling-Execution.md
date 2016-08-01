You can control execution of multiple bash commands with a single ahoy command like so:

* `&&` means continue if success (exit 0)
* `;` means always run the next command
* `||` means to run the next command if the previous one failed. (exit non-zero)

Example:

```Yaml
...
  commands:
    only-run-1-2-and-4:
      cmd: |
        echo "1 - If this passes" &&
        echo "2 - Then do this" ||
        echo "3 - Or do this if 1 or 2 fails (returns non-zero)" ;
        echo "4 - Do this no matter what"
```