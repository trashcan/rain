#!/usr/bin/env bash
if [[ -d "/usr/local/etc/bash_completion.d/" ]]; then
    cp -v rain.bash /usr/local/etc/bash_completion.d/rain
elif [[ -d "/etc/bash_completion.d/" ]]; then
    cp -v rain.bash /etc/bash_completion.d
fi
