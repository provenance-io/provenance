#!/bin/bash -e

##########################################
##                                      ##
##  This script downloads the provwasm  ##
##  tutorial smart contract to be used  ##
##  in sims tests.                      ##
##                                      ##
##########################################

export Provwasm_version="v1.1.0"


wget "https://github.com/provenance-io/provwasm/releases/download/$Provwasm_version/tutorial_contract.zip"

unzip "tutorial_contract.zip"
rm tutorial_contract.zip
mv -f ./provwasm_tutorial.wasm ./app/sim_contracts/tutorial.wasm
