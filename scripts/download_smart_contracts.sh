#!/bin/bash -e

##########################################
##                                      ##
##  This script downloads the provwasm  ##
##  tutorial smart contract to be used  ##
##  in sims tests.                      ##
##                                      ##
##########################################

export Provwasm_version="v1.0.0-beta3"

wget "https://github.com/provenance-io/provwasm/releases/download/$Provwasm_version/provwasm_tutorial.zip"

unzip "provwasm_tutorial.zip"
rm provwasm_tutorial.zip
mv -f ./provwasm_tutorial.wasm ./app/sim_contracts/tutorial.wasm
