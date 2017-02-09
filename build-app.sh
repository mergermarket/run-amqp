#!/bin/bash -e

set -o errexit
set -o nounset
set -o pipefail

if [ ! $(command -v gometalinter) ]
then
	go get github.com/alecthomas/gometalinter
	gometalinter --install --vendor
fi

gometalinter \
    --vendor \
	--exclude='error return value not checked.*(Close|Log|Print).*\(errcheck\)$' \
	--exclude='.*_test\.go:.*error return value not checked.*\(errcheck\)$' \
	--exclude='duplicate of.*_test.go.*\(dupl\)$' \
	--disable=aligncheck \
	--disable=gotype \
	--disable=unconvert \
	--disable=aligncheck \
	--disable=gas \
	--cyclo-over=20 \
	--tests \
	--deadline=20s

go fmt $(go list ./... | grep -v /vendor/)

 for pkg in $(go list ./... | grep -v /vendor/)
    do
      echo "pkg=$pkg"
      go test -coverprofile=coverage/coverage.out -covermode=count $pkg
      if [ -f coverage/coverage.out ]; then
        tail -n +2 coverage/coverage.out >> coverage/coverage-all.out
      fi
    done