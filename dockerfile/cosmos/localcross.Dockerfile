# This file exists to override Heighliner's embedded "cosmos/localcross.Dockerfile"
# template, which hardcodes `FROM ghcr.io/strangelove-ventures/infra-toolkit:v0.1.12`.
# That image now 403s on an anonymous pull because the strangelove-ventures GitHub
# org (and its ghcr.io packages) moved to amygdala-labs; the package now lives at
# ghcr.io/amygdala-labs/infra-toolkit:v0.1.12. Heighliner's released template hasn't
# been updated to point at the new location, so the embedded copy fails every build.
# This is otherwise a verbatim copy of heighliner v1.7.6's embedded template with
# only that one FROM line changed. Once upstream fixes it, this override (and its
# directory) can be deleted.
#
# It must live at exactly this path: heighliner looks for a local override at
# "<repo-root>/dockerfile/<type>/<name>.Dockerfile" (see dockerfileEmbeddedOrLocal
# in github.com/amygdala-labs/heighliner's builder/builder.go) before falling back
# to its embedded copy. The "dockerfile" directory name is hardcoded in heighliner
# itself, not configurable via the action's inputs, so it can't be moved or renamed.
#
# TODO: https://github.com/provenance-io/provenance/issues/2835 Check if this file can be deleted.
ARG BASE_VERSION
FROM --platform=$BUILDPLATFORM golang:${BASE_VERSION} AS build-env

RUN apk add --update --no-cache curl make git libc-dev bash gcc linux-headers eudev-dev

ARG TARGETARCH
ARG BUILDARCH

RUN if [ "${TARGETARCH}" = "arm64" ] && [ "${BUILDARCH}" != "arm64" ]; then\
        wget -c https://storage.googleapis.com/strangelove-public/musl/aarch64-linux-musl-cross.tgz -O - | tar -xzvv --strip-components 1 -C /usr;\
    elif [ "${TARGETARCH}" = "amd64" ] && [ "${BUILDARCH}" != "amd64" ]; then\
        wget -c https://storage.googleapis.com/strangelove-public/musl/x86_64-linux-musl-cross.tgz -O - | tar -xzvv --strip-components 1 -C /usr;\
    fi

ARG GITHUB_ORGANIZATION
ARG REPO_HOST

WORKDIR /go/src/${REPO_HOST}/${GITHUB_ORGANIZATION}/${GITHUB_REPO}

ARG GITHUB_REPO
ARG VERSION
ARG BUILD_TIMESTAMP

ADD . .

ARG BUILD_TARGET
ARG BUILD_ENV
ARG BUILD_TAGS
ARG PRE_BUILD
ARG BUILD_DIR
ARG WASMVM_VERSION

ARG CLONE_KEY

RUN if [ ! -z "${CLONE_KEY}" ]; then\
  mkdir -p ~/.ssh;\
  echo "${CLONE_KEY}" | base64 -d > ~/.ssh/id_ed25519;\
  chmod 600 ~/.ssh/id_ed25519;\
  apk add openssh;\
  git config --global --add url."ssh://git@github.com/".insteadOf "https://github.com/";\
  ssh-keyscan github.com >> ~/.ssh/known_hosts;\
  fi

RUN set -eux;\
    LIBDIR=/lib;\
    if [ "${TARGETARCH}" = "arm64" ]; then\
      export ARCH=aarch64;\
      if [ "${BUILDARCH}" != "arm64" ]; then\
        LIBDIR=/usr/aarch64-linux-musl/lib;\
        mkdir -p $LIBDIR;\
        export CC=aarch64-linux-musl-gcc CXX=aarch64-linux-musl-g++;\
      fi;\
    elif [ "${TARGETARCH}" = "amd64" ]; then\
      export ARCH=x86_64;\
      if [ "${BUILDARCH}" != "amd64" ]; then\
        LIBDIR=/usr/x86_64-linux-musl/lib;\
        mkdir -p $LIBDIR;\
        export CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++;\
      fi;\
    fi;\
    if [ ! -z "${WASMVM_VERSION}" ]; then\
      WASMVM_REPO=$(echo $WASMVM_VERSION | awk '{print $1}');\
      WASMVM_VERS=$(echo $WASMVM_VERSION | awk '{print $2}');\
      wget -O $LIBDIR/libwasmvm_muslc.a https://${WASMVM_REPO}/releases/download/${WASMVM_VERS}/libwasmvm_muslc.${ARCH}.a;\
      ln $LIBDIR/libwasmvm_muslc.a $LIBDIR/libwasmvm_muslc.$(uname -m).a;\
    fi;\
    export GOOS=linux GOARCH=$TARGETARCH CGO_ENABLED=1 LDFLAGS='-linkmode external -extldflags "-static"';\
    if [ ! -z "$PRE_BUILD" ]; then sh -c "${PRE_BUILD}"; fi;\
    if [ ! -z "$BUILD_TARGET" ]; then\
      if [ ! -z "$BUILD_ENV" ]; then export ${BUILD_ENV}; fi;\
      if [ ! -z "$BUILD_TAGS" ]; then export "${BUILD_TAGS}"; fi;\
      if [ ! -z "$BUILD_DIR" ]; then cd "${BUILD_DIR}"; fi;\
      sh -c "${BUILD_TARGET}";\
    fi

RUN if [ -d "/go/bin/linux_${TARGETARCH}" ]; then mv /go/bin/linux_${TARGETARCH}/* /go/bin/; fi

# Copy all binaries to /root/bin, for a single place to copy into final image.
# If a colon (:) delimiter is present, binary will be renamed to the text after the delimiter.
RUN mkdir /root/bin
ARG RACE
ARG BINARIES
ENV BINARIES_ENV ${BINARIES}
RUN bash -c 'set -eux;\
  BINARIES_ARR=();\
  IFS=, read -ra BINARIES_ARR <<< "$BINARIES_ENV";\
  for BINARY in "${BINARIES_ARR[@]}"; do\
    BINSPLIT=();\
    IFS=: read -ra BINSPLIT <<< "$BINARY";\
    BINPATH=${BINSPLIT[1]+"${BINSPLIT[1]}"};\
    BIN="$(eval "echo "${BINSPLIT[0]+"${BINSPLIT[0]}"}"")";\
    if [ ! -z "$RACE" ] && GOVERSIONOUT=$(go version -m $BIN); then\
      if echo $GOVERSIONOUT | grep build | grep "-race=true"; then\
        echo "Race detection is enabled in binary";\
      else\
        echo "Race detection not enabled in binary!";\
        exit 1;\
      fi;\
    fi;\
    if [ ! -z "$BINPATH" ]; then\
      if [[ $BINPATH == *"/"* ]]; then\
        mkdir -p "$(dirname "${BINPATH}")";\
        cp "$BIN" "${BINPATH}";\
      else\
        cp "$BIN" "/root/bin/${BINPATH}";\
      fi;\
    else\
      cp "$BIN" /root/bin/;\
    fi;\
  done'

RUN mkdir -p /root/lib
ARG LIBRARIES
ENV LIBRARIES_ENV ${LIBRARIES}
RUN bash -c 'set -eux;\
  LIBRARIES_ARR=($LIBRARIES_ENV); for LIBRARY in "${LIBRARIES_ARR[@]}"; do cp $LIBRARY /root/lib/; done'

# Copy over directories
RUN mkdir -p /root/dir_abs && touch /root/dir_abs.list
ARG DIRECTORIES
ENV DIRECTORIES_ENV ${DIRECTORIES}
RUN bash -c 'set -eux;\
  DIRECTORIES_ARR=($DIRECTORIES_ENV);\
  i=0;\
  for DIRECTORY in "${DIRECTORIES_ARR[@]}"; do \
    cp -R $DIRECTORY /root/dir_abs/$i;\
    echo $DIRECTORY >> /root/dir_abs.list;\
    ((i = i + 1));\
  done'

# Use minimal busybox from infra-toolkit image for final scratch image
# NOTE: repointed from ghcr.io/strangelove-ventures/infra-toolkit (403s anonymously
# since the org moved) to the same image under its new home, ghcr.io/amygdala-labs.
FROM ghcr.io/amygdala-labs/infra-toolkit:v0.1.12 AS infra-toolkit
RUN addgroup --gid 1025 -S heighliner && adduser --uid 1025 -S heighliner -G heighliner

# Use alpine to source the latest CA certificates
FROM alpine:3 as alpine-3

# Build final image from scratch
FROM scratch

LABEL org.opencontainers.image.source="https://github.com/strangelove-ventures/heighliner"

WORKDIR /bin

# Install minimal busybox as `sh` and `ln` binaries
# sh allows using `RUN` commands
COPY --from=infra-toolkit /busybox/busybox /bin/sh
# ln creates hardlinks for exposed binaries from infra-toolkit min config
COPY --from=infra-toolkit /busybox/busybox /bin/ln

# Install jq
COPY --from=infra-toolkit /usr/local/bin/jq /bin/

# Add hard links for utils
# Will then only have one copy of the busybox minimal binary file with all utils pointing to the same underlying inode
RUN for b in \
  cat \
  date \
  df \
  dirname \
  du \
  env \
  grep \
  head \
  less \
  ls \
  md5sum \
  mkdir \
  mv \
  pwd \
  rm \
  sed \
  sha1sum \
  sha256sum \
  sha3sum \
  sha512sum \
  sleep \
  stty \
  tail \
  tar \
  tee \
  tr \
  vi \
  watch \
  which \
  ; do ln ln $b; done; \
  rm -rf sh; \
  ln ln sh;

# Copy over absolute path directories
COPY --from=build-env /root/dir_abs /root/dir_abs
COPY --from=build-env /root/dir_abs.list /root/dir_abs.list

# Move absolute path directories to their absolute locations.
RUN sh -c 'i=0; while read DIR; do\
      echo "$i: $DIR";\
      PLACEDIR="$(dirname "$DIR")";\
      mkdir -p "$PLACEDIR";\
      mv /root/dir_abs/$i $DIR;\
      i=$((i+1));\
    done < /root/dir_abs.list'

# Install chain binaries
COPY --from=build-env /root/bin /bin

# Install libraries
COPY --from=build-env /root/lib /lib

# Install trusted CA certificates
COPY --from=alpine-3 /etc/ssl/cert.pem /etc/ssl/cert.pem

# Install heighliner user
COPY --from=infra-toolkit /etc/passwd /etc/passwd
COPY --from=infra-toolkit --chown=1025:1025 /home/heighliner /home/heighliner
COPY --from=infra-toolkit --chown=1025:1025 /tmp /tmp

WORKDIR /home/heighliner
USER heighliner
