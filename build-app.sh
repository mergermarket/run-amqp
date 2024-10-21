#!/bin/bash -e

set -o errexit
set -o nounset
set -o pipefail

go fmt $(go list ./... | grep -v /vendor/)
go test $(go list ./... | grep -v acceptance-tests ) --cover -timeout 25s