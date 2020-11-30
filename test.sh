#!/bin/bash

set -eo pipefail

trap './dist/linux_linux_amd64/reporter -debug true' EXIT
go test -race -count=1 -v 2>&1 ./... | go-junit-report > report.xml
