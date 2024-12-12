alias ah='ahoy'

# Exit early if ahoy plugin is already loaded.
if (( ${+commands[ahoy]} )); then
    return
fi

# If the completion file doesn't exist yet, we need to autoload it and
# bind it to `ahoy`. Otherwise, compinit will have already done that.
if [[ ! -f "$ZSH_CACHE_DIR/completions/_ahoy" ]]; then
    typeset -g -A _comps
    autoload -Uz _ahoy
    _comps[ahoy]=_ahoy
fi

# Save completions file in cache directory. 
cp "${0:h}/_ahoy" "$ZSH_CACHE_DIR/completions/_ahoy"
