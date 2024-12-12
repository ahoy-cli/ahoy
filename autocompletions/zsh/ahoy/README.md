# Ahoy Oh-My-Zsh plugin

1. Clone the zsh-ahoy repository into `$ZSH_CUSTOM/plugins` (by default `~/.oh-my-zsh/custom/plugins`)

```bash
git clone https://github.com/ahoy-cli/zsh-ahoy ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/ahoy
```

2. Add the ahoy zsh plugin to the list of plugins for Oh My Zsh to load, inside `~/.zshrc`:

```
plugins=(
    # Other plugins above...
    ahoy
)
```

3. Start a new terminal session.
