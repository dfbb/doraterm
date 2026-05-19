# add dsh to path, source dynamic script from dsh token
DORATERM_DSHBINDIR={{.WSHBINDIR}}
export PATH="$DORATERM_DSHBINDIR:$PATH"
source <(dsh token "$DORATERM_SWAPTOKEN" zsh 2>/dev/null)
unset DORATERM_SWAPTOKEN

# Source the original zshrc only if ZDOTDIR has not been changed
if [ "$ZDOTDIR" = "$DORATERM_ZDOTDIR" ]; then
  [ -f ~/.zshrc ] && source ~/.zshrc
fi

if [[ ":$PATH:" != *":$DORATERM_DSHBINDIR:"* ]]; then
  export PATH="$DORATERM_DSHBINDIR:$PATH"
fi
unset DORATERM_DSHBINDIR

if [[ -n ${_comps+x} ]]; then
  source <(dsh completion zsh)
fi

# fix history (macos)
if [[ "$HISTFILE" == "$DORATERM_ZDOTDIR/.zsh_history" ]]; then
  HISTFILE="$HOME/.zsh_history"
fi

typeset -g _DORATERM_SI_FIRSTPRECMD=1

# shell integration
_doraterm_si_blocked() {
  [[ -n "$TMUX" || -n "$STY" || "$TERM" == tmux* || "$TERM" == screen* ]]
}

_doraterm_si_urlencode() {
  if (( $+functions[omz_urlencode] )); then
    omz_urlencode "$1"
  else
    local s="$1"
    # Escape % first
    s=${s//\%/%25}
    # Common reserved characters in file paths
    s=${s//\ /%20}
    s=${s//\#/%23}
    s=${s//\?/%3F}
    s=${s//\&/%26}
    s=${s//\;/%3B}
    s=${s//\+/%2B}
    printf '%s' "$s"
  fi
}

_doraterm_si_compmode() {
  # fzf-based completion wins
  if typeset -f _fzf_tab_complete >/dev/null 2>&1 || typeset -f _fzf_complete >/dev/null 2>&1; then
    echo "fzf"
    return
  fi

  # Check zstyle menu setting
  local _menuval
  if zstyle -s ':completion:*' menu _menuval 2>/dev/null; then
    if [[ "$_menuval" == *select* ]]; then
      echo "menu-select"
    else
      echo "menu"
    fi
    return
  fi

  echo "standard"
}

_doraterm_si_osc7() {
  _doraterm_si_blocked && return
  local encoded_pwd=$(_doraterm_si_urlencode "$PWD")
  printf '\033]7;file://localhost%s\007' "$encoded_pwd"  # OSC 7 - current directory
}

_doraterm_si_precmd() {
  local _doraterm_si_status=$?
  _doraterm_si_blocked && return
  # D;status for previous command (skip before first prompt)
  if (( !_DORATERM_SI_FIRSTPRECMD )); then
    printf '\033]16162;D;{"exitcode":%d}\007' "$_doraterm_si_status"
  else
    local uname_info=$(uname -smr 2>/dev/null)
    local omz=false
    local comp=$(_doraterm_si_compmode)
    [[ -n "$ZSH" && -r "$ZSH/oh-my-zsh.sh" ]] && omz=true
    printf '\033]16162;M;{"shell":"zsh","shellversion":"%s","uname":"%s","integration":true,"omz":%s,"comp":"%s"}\007' "$ZSH_VERSION" "$uname_info" "$omz" "$comp"
    # OSC 7 only sent on first prompt - chpwd hook handles directory changes
    _doraterm_si_osc7
  fi
  printf '\033]16162;A\007'
  _DORATERM_SI_FIRSTPRECMD=0
}

_doraterm_si_preexec() {
  _doraterm_si_blocked && return
  local cmd="$1"
  local cmd_length=${#cmd}
  if [ "$cmd_length" -gt 8192 ]; then
    cmd=$(printf '# command too large (%d bytes)' "$cmd_length")
  fi
  local cmd64
  cmd64=$(printf '%s' "$cmd" | base64 2>/dev/null | tr -d '\n\r')
  if [ -n "$cmd64" ]; then
    printf '\033]16162;C;{"cmd64":"%s"}\007' "$cmd64"
  else
    printf '\033]16162;C\007'
  fi
}

typeset -g DORATERM_SI_INPUTEMPTY=1

_doraterm_si_inputempty() {
  _doraterm_si_blocked && return
  
  local current_empty=1
  if [[ -n "$BUFFER" ]]; then
    current_empty=0
  fi
  
  if (( current_empty != DORATERM_SI_INPUTEMPTY )); then
    DORATERM_SI_INPUTEMPTY=$current_empty
    if (( current_empty )); then
      printf '\033]16162;I;{"inputempty":true}\007'
    else
      printf '\033]16162;I;{"inputempty":false}\007'
    fi
  fi
}

autoload -Uz add-zle-hook-widget 2>/dev/null
if (( $+functions[add-zle-hook-widget] )); then
  add-zle-hook-widget zle-line-init _doraterm_si_inputempty
  add-zle-hook-widget zle-line-pre-redraw _doraterm_si_inputempty
fi

autoload -U add-zsh-hook
add-zsh-hook precmd  _doraterm_si_precmd
add-zsh-hook preexec _doraterm_si_preexec
add-zsh-hook chpwd   _doraterm_si_osc7