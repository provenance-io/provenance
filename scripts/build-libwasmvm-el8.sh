#!/usr/bin/env bash
# This script builds libwasmvm.x86_64.so in a Rocky Linux 8 container so that it only requires GLIBC 2.28.
# It's for validators running RHEL 8.x (and other EL8 distros).
#
# Background: The prebuilt libwasmvm.x86_64.so from wasmvm v3.0.8 requires GLIBC 2.30 (v3.0.7 only needed 2.28).
# So nodes on RHEL 8.x (GLIBC 2.28) can't load it.
#
# The result is built from the exact same wasmvm source (and Cargo.lock) that's in the go module,
# using the same Rust version and the same cargo build command that upstream uses (plus --locked).
# The only difference is the GLIBC it's linked against, which doesn't affect contract execution.
#
# The library only depends on the wasmvm version, not the provenance version. So the result can be used
# with any provenanced built with the same wasmvm version.
#
# Usage: scripts/build-libwasmvm-el8.sh [<output dir>]
# Default output dir is build/libwasmvm-el8.
# Requires docker and go. Must be run from the repo root.
set -euo pipefail

max_glibc='2.28'
wasmvm_mod='github.com/CosmWasm/wasmvm/v3'

out_dir="${1:-build/libwasmvm-el8}"
mkdir -p "$out_dir"
out_dir="$( cd "$out_dir" && pwd )"

go mod download "$wasmvm_mod"
# The -mod=mod is needed so that the module cache dir is provided even if there's a vendor/ dir.
wasmvm_version="$( go list -mod=mod -m -f '{{.Version}}' "$wasmvm_mod" )"
wasmvm_dir="$( go list -mod=mod -m -f '{{.Dir}}' "$wasmvm_mod" )"

# Use the same Rust version that upstream used to build this version of wasmvm.
rust_version="$( grep -Eo -- '--default-toolchain [0-9.]+' "$wasmvm_dir/builders/Dockerfile.debian" | awk '{print $2}' )"
if [[ -z "$rust_version" ]]; then
  echo "Could not identify the Rust version from $wasmvm_dir/builders/Dockerfile.debian" >&2
  exit 1
fi

echo "wasmvm version: $wasmvm_version"
echo "Rust version: $rust_version"
echo "Output dir: $out_dir"

# The module cache is read-only, so the source is copied into the container before building.
docker run --rm --platform linux/amd64 \
  -v "$wasmvm_dir/libwasmvm:/src:ro" \
  -v "$out_dir:/out" \
  -e RUST_VERSION="$rust_version" \
  -e MAX_GLIBC="$max_glibc" \
  rockylinux:8 \
  bash -euo pipefail -c '
    dnf install -y gcc make clang clang-devel llvm-devel binutils
    curl --proto "=https" --tlsv1.2 -sSf -o /tmp/rustup-init "https://static.rust-lang.org/rustup/dist/x86_64-unknown-linux-gnu/rustup-init"
    chmod +x /tmp/rustup-init
    /tmp/rustup-init -y --no-modify-path --profile minimal --default-toolchain "$RUST_VERSION"
    export PATH="$HOME/.cargo/bin:$PATH"
    rustc --version
    cargo --version
    ldd --version | sed -n 1p

    cp -r /src /build
    cd /build
    export CARGO_REGISTRIES_CRATES_IO_PROTOCOL=sparse
    export CC=clang
    export CXX=clang++
    cargo build --release --locked --target x86_64-unknown-linux-gnu
    cp target/x86_64-unknown-linux-gnu/release/libwasmvm.so /out/libwasmvm.x86_64.so

    glibc="$( objdump -T /out/libwasmvm.x86_64.so | grep -Eo "GLIBC_[0-9.]+" | sed "s/^GLIBC_//" | sort -Vu | tail -n 1 )"
    echo "Highest GLIBC required: $glibc"
    if [[ "$( printf "%s\n%s\n" "$glibc" "$MAX_GLIBC" | sort -V | tail -n 1 )" != "$MAX_GLIBC" ]]; then
      echo "Error: GLIBC $glibc is newer than $MAX_GLIBC." >&2
      exit 1
    fi
  '

( cd "$out_dir" && shasum -a 256 libwasmvm.x86_64.so > libwasmvm.x86_64.so.sha256 && cat libwasmvm.x86_64.so.sha256 )
echo "Done: $out_dir/libwasmvm.x86_64.so (wasmvm $wasmvm_version, Rust $rust_version, rockylinux:8)"
