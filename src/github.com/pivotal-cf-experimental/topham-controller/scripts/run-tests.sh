#!/usr/bin/env bash

set -euo pipefail

pushd ${GOPATH}/src/github.com/starkandwayne/eden
  rm -rf eden
  go build .
popd

pushd $(dirname $0)
  source ../.envrc
  mv ${GOPATH}/src/github.com/starkandwayne/eden/eden eden-test
  chmod +x eden-test
  export CLI_BINARY_PATH=$(dirname `pwd`)/$(basename `pwd`)/eden-test

  pushd ..
  ginkgo -v -r system
  popd

  rm eden-test
popd
