#!/usr/bin/env bash
set -e

if [ "$1" == "" ]; then
	echo "Usage: $0 <output-file>"
	exit 1
fi

PROTO_DIR=proto

dir="$(pwd)"
zip="${1}"

rm -f "${zip}"

# Colorize the ouput.
red='\e[0;31m'
green='\e[0;32m'
lite_blue='\e[1;34m'
lite_red='\e[1;31m'
yellow='\e[1;33m'
off='\e[0m'

# Include all third_party protos in the final zipball.
cd "${dir}/third_party" || exit 1
find "." -name \*.proto -print0 | while read -rd $'\0' d; do
	echo -en " * Adding ${lite_red}external${off} proto ${yellow}${d}${off} ... "
    if find "${d}" -name \*.proto | zip "${zip}" -@ >/dev/null; then
		  echo -e "[${green}OK${off}]"
	  else
		  echo -e "[${red}!!${off}]"
	fi
done

# Include all provenance protos in the final zipball.
cd "${dir}" || exit 1
find "${PROTO_DIR}" -name \*.proto -print0 | while read -rd $'\0' d; do
	echo -en " * Adding ${lite_blue}internal${off} protos ${yellow}${d}${off} ... "
	if find "${d}" -name \*.proto | zip "${zip}" -@ >/dev/null; then
    echo -e "[${green}OK${off}]"
  else
    echo -e "[${red}!!${off}]"
  fi
done

# Formatted for gh workflow action
echo
echo "::set-output name=protos::$zip"