# fish completion for pixelize.
#
# Install: copy this file to your fish completions directory, e.g.
#     ~/.config/fish/completions/pixelize.fish

function __pixelize_presets
    pixelize --list-presets 2>/dev/null | awk '{print $1}'
    echo all
end

function __pixelize_palettes
    pixelize --list-palettes 2>/dev/null | awk '{print $1}'
end

# Commands
complete -c pixelize -n '__fish_use_subcommand' -a dump-all -d 'Copy built-in presets and palettes to your config dir'

# Flags that take a value
complete -c pixelize -l block -d 'pixel block size' -x
complete -c pixelize -l scale -d 'nearest-neighbor output scale' -x
complete -c pixelize -l colors -d 'reduce palette to N colors' -x
complete -c pixelize -l mode -d 'sampling mode' -x -a 'average median dominant'
complete -c pixelize -l palette -d 'color palette' -x -a '(__pixelize_palettes)'
complete -c pixelize -l preset -d 'apply a named preset' -x -a '(__pixelize_presets)'
complete -c pixelize -l palette-out -d 'write the output palette to a JSON file' -r

# Boolean flags
complete -c pixelize -l dither -d 'apply Floyd-Steinberg dithering'
complete -c pixelize -l bw -d 'reduce to black and white'
complete -c pixelize -l transparent -d 'make white pixels transparent'
complete -c pixelize -l sharpen -d 'apply edge-enhancement sharpening'
complete -c pixelize -l grid -d 'draw a subtle pixel-grid overlay'
complete -c pixelize -l logical -d 'export logical pixel grid without scaling'
complete -c pixelize -l list-presets -d 'list available presets and exit'
complete -c pixelize -l list-presets-json -d 'list available presets as JSON and exit'
complete -c pixelize -l list-palettes -d 'list available palettes and exit'
complete -c pixelize -l force -d 'overwrite existing files (used with dump-all)'
complete -c pixelize -l version -d 'print version and exit'
complete -c pixelize -l help -d 'print usage and exit'
