#!/bin/bash

SED=$(command -v gsed || command -v sed)

function setmgo() {
    target="$1"
    find ./vendor/github.com/tsuru -name "*.go" -exec ${SED} -i'' 's|"github.com/globalsign/mgo|"'"$target"'|g' {} +
    find ./vendor/github.com/tsuru -name "*.go" -exec ${SED} -i'' 's|"gopkg.in/mgo.v2|"'"$target"'|g' {} +
    ${SED} -i'' 's|mgo "github.com/globalsign/mgo|mgo "'"$target"'|g' main.go
    ${SED} -i'' 's|mgo "gopkg.in/mgo.v2|mgo "'"$target"'|g' main.go
}

function runtest() {
    target="$1"
    setmgo "$target"
    failurecount=0
    tries=30
    echo "RUNNING: $target - n=$tries"
    go build -i
    for i in $(seq $tries); do
        ./mgostress 2>/dev/null
        exitcode="$?"
        if [[ "$exitcode" != "0" ]]; then
            ((failurecount++))
        fi
    done
    result=$(echo "scale=2; (${failurecount}/${tries})*100" | bc)
    echo "TARGET $target - failures: ${result}%"
}

target="$1"
if [[ "$target" == "" ]]; then
    runtest "gopkg.in/mgo.v2"
    runtest "github.com/globalsign/mgo"
else
    runtest "$target"
fi
