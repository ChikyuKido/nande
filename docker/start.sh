#!/bin/sh
cd extensions
./build-extensions.sh
cd ..
exec ./nande run
