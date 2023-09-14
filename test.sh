#!/bin/bash

set -eo pipefail

trap './dist/linux_linux_amd64_v1/reporter -debug true' EXIT
go test -race -count=1 -v 2>&1 ./... | go-junit-report -iocopy -set-exit-code -out report.xml