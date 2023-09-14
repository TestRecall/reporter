#!/bin/bash

set -eo pipefail

trap './dist/linux_linux_amd64_v1/reporter -debug true' EXIT
go test -race -count=1 -v 2>&1 ./... | tee /dev/tty | go-junit-report -set-exit-code > report.xml
