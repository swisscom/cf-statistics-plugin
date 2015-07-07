#!/bin/bash

cf uninstall-plugin Statistics
rm -f $GOPATH/bin/cf-statistics-plugin

set -e
go build
go install

cf install-plugin $GOPATH/bin/cf-statistics-plugin
cf plugins
