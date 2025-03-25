#!/bin/sh

#if the extensions dir is empty move the default extensions into it
if [ ! "$(ls -A /app/extensions)" ]; then
    cp -r /app/default-extensions/* /app/extensions/
    echo "Default extensions copied to /app/extensions"
fi

exec ./nande run
