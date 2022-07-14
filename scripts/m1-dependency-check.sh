#!/bin/bash

# Install dependency with brew
# Requires the name of the package to install
has_dependency() {
    if brew list $1 &>/dev/null; then
        return 1
    else
        return 0
    fi
}

# Ensure homebrew is installed
if [[ $(command -v brew) == "" ]]; then
    echo "This script requires Homebrew. Please install and run again..."
    exit 1
fi

# Check dependencies
kafka_dependencies=( librdkafka openssl pkg-config )
for dependency in "${kafka_dependencies[@]}"
do
    has_dependency $dependency
    status=$?
    if [[ $status == 0 ]]; then
        echo "Missing $dependency. Recommended to run: make librdkafka"
        exit 1
    fi
done

# Check that PKG_CONFIG_PATH is set
case ":$PKG_CONFIG_PATH:" in
  *:"$( brew --prefix openssl )/lib/pkgconfig":*) : ;;
  *) echo "PKG_CONFIG_PATH is missing openssl. Recommended to run: export PKG_CONFIG_PATH=\"\$( brew --prefix openssl )\"/lib/pkgconfig\"\${PKG_CONFIG_PATH:+:\$PKG_CONFIG_PATH}\""; exit 1;
esac
exit 0