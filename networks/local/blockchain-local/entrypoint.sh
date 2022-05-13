#!/usr/bin/env bash

##
## Input parameters
##
ID=${ID:-0}
export PIO_HOME="/provenance/node${ID}"

# This is for M1
export PKG_CONFIG_PATH="/usr/local/lib"
export LD_LIBRARY_PATH="/usr/local/lib"

BINARY=/usr/bin/${BINARY:-provenanced}
LOG=${LOG:-provenance.log}

##
## Run binary with all parameters
##
if [ -d "$(dirname "${PIO_HOME}"/"${LOG}")" ]; then
  "${BINARY}" -t --home "${PIO_HOME}" "$@" | tee "${PIO_HOME}/${LOG}"
else
  "${BINARY}" -t --home "${PIO_HOME}" "$@" 
fi

