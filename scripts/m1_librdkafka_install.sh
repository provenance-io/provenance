#!/bin/bash

# Install dependency with brew
# Requires the name of the package to install
brew_install() {
    if brew list $1 &>/dev/null; then
        echo "Dependency $1 already installed. Skipping..."
    else
        echo "Installing $1..."
        brew install $1
    fi
}

# Ensure homebrew is installed
if [[ $(command -v brew) == "" ]]; then
    echo "This script requires Homebrew. Please install and run again..."
    exit 0
fi

# Install dependencies
brew_install "librdkafka"
brew_install "openssl"
brew_install "pkg-config"

# Check that PKG_CONFIG_PATH is set
case ":$PKG_CONFIG_PATH:" in
  *:"$( brew --prefix openssl )/lib/pkgconfig":*) : ;;
  *) echo ""; echo "openssl is MISSING from PKG_CONFIG_PATH. Set it with the following command:"; echo 'export PKG_CONFIG_PATH="$( brew --prefix openssl )"/lib/pkgconfig"${PKG_CONFIG_PATH:+:$PKG_CONFIG_PATH}"' ;;
esac
echo ""
echo "Installation complete!"