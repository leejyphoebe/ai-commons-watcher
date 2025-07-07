#!/bin/bash

repo="git@github.com:googlercolin/slimsc.git"
dir="/scratch/users/ntu/$USER/slimsc"

if ! git clone "${repo}" "${dir}" 2>/dev/null && [ -d "${dir}" ] ; then
    echo "Clone failed because the folder ${dir} exists"
fi

# explicitly create the logs directory
mkdir -p "logs"
