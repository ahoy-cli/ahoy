# Ahoy Oh-My-Zsh plugin

1. Copy this folder into the `$ZSH_CUSTOM/plugins` directory (by default `~/.oh-my-zsh/custom/plugins`). The command below is being run from the perspective of this directory.

```bash
cp -r ../ahoy ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/ahoy
```

2. Add the ahoy zsh plugin to the list of plugins for Oh My Zsh to load, inside `~/.zshrc`:

```
plugins=(
    # Other plugins above...
    ahoy
)
```

3. Start a new terminal session.
