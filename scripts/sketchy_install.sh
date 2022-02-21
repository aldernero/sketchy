#!/usr/bin/env bash

if [ $# -eq 0 ]
then
    echo "Usage: sketchy_install.sh target_directory"
    exit 1
fi

DIR=$1

if [ -d "$DIR" ]
then
    echo "$DIR already exists"
    exit 1
fi

mkdir "$DIR"

cp -r ../template "$DIR"/
go build -o "$DIR"/sketchy ../cmd/sketchy/main.go

echo "Sucessfully installed sketchy environment to $DIR"