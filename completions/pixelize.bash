# bash completion for pixelize.
#
# Install: source this file from your ~/.bashrc, e.g.
#     source /path/to/pixelize.bash
# or copy it into your bash-completion directory, e.g.
#     /usr/local/etc/bash_completion.d/pixelize   (macOS/Homebrew)
#     /etc/bash_completion.d/pixelize             (Linux)

_pixelize() {
    local cur prev words cword
    _init_completion 2>/dev/null || {
        cur="${COMP_WORDS[COMP_CWORD]}"
        prev="${COMP_WORDS[COMP_CWORD-1]}"
    }

    local flags="--block --scale --colors --dither --bw --transparent \
--sharpen --grid --logical --mode --palette --preset --palette-out \
--list-presets --list-presets-json --list-palettes --force --version --help"

    # Complete the value for flags that take an argument.
    case "$prev" in
        --mode)
            COMPREPLY=( $(compgen -W "average median dominant" -- "$cur") )
            return
            ;;
        --preset)
            local presets
            presets="$(pixelize --list-presets 2>/dev/null | awk '{print $1}') all"
            COMPREPLY=( $(compgen -W "$presets" -- "$cur") )
            return
            ;;
        --palette)
            local palettes
            palettes="$(pixelize --list-palettes 2>/dev/null | awk '{print $1}')"
            COMPREPLY=( $(compgen -W "$palettes" -- "$cur") )
            return
            ;;
        --palette-out)
            COMPREPLY=( $(compgen -f -X '!*.json' -- "$cur") )
            return
            ;;
        --block|--scale|--colors)
            return
            ;;
    esac

    if [[ "$cur" == -* ]]; then
        COMPREPLY=( $(compgen -W "$flags" -- "$cur") )
        return
    fi

    # First positional: offer the dump-all command plus files.
    if [[ $COMP_CWORD -eq 1 ]]; then
        COMPREPLY=( $(compgen -W "dump-all" -- "$cur") $(compgen -f -- "$cur") )
        return
    fi

    COMPREPLY=( $(compgen -f -- "$cur") )
}

complete -F _pixelize pixelize
