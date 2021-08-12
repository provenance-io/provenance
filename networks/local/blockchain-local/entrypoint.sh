#!/usr/bin/env bash

##
## Input parameters
##
ID=${ID:-0}
export PIO_HOME="/provenance/node${ID}"
BINARY=/usr/bin/${BINARY:-provenanced}
LOG=${LOG:-provenance.log}

##
## Run binary with all parameters
##
if [ -d "$(dirname "${PIO_HOME}"/"${LOG}")" ]; then
  echo "Look here!!!"
  echo "using log"
  echo "$@"
  "${BINARY}" -t --home "${PIO_HOME}" "$@" | tee "${PIO_HOME}/${LOG}"
else
  echo "Look here!!!"
  echo "${BINARY}"
  echo "$@"
  "${BINARY}" -t --home "${PIO_HOME}" "$@" 
fi

