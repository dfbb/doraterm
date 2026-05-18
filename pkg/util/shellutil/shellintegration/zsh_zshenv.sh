# Store the initial ZDOTDIR value
DORATERM_ZDOTDIR="$ZDOTDIR"

# Source the original zshenv
[ -f ~/.zshenv ] && source ~/.zshenv

# Detect if ZDOTDIR has changed
if [ "$ZDOTDIR" != "$DORATERM_ZDOTDIR" ]; then
  # If changed, manually source your custom zshrc from the original DORATERM_ZDOTDIR
  [ -f "$DORATERM_ZDOTDIR/.zshrc" ] && source "$DORATERM_ZDOTDIR/.zshrc"
fi