ARG GO_VERSION=v1.21.1

FROM --platform=$BUILDPLATFORM goreleaser/goreleaser-cross-base:${GO_VERSION}

WORKDIR /usr/src/app

ENV DEBIAN_FRONTEND=noninteractive

# Install Libgtk, webkit and NSIS
RUN dpkg --add-architecture amd64 \
  && apt-get -qq update \
  && apt-get -qq install -y libgtk-3-dev:amd64 libwebkit2gtk-4.0-dev:amd64

RUN dpkg --add-architecture arm64 \
  && apt-get -qq update \
  && apt-get -qq install -y libgtk-3-dev:arm64 libwebkit2gtk-4.0-dev:arm64

# NSIS is only needed for Windows, so we install the one that matches the build platform
RUN apt-get -qq install -y nsis

ARG NODE_MAJOR_VERSION=18

# Install NodeJS
RUN apt-get -qq install -y ca-certificates curl gnupg && \
    mkdir -p /etc/apt/keyrings && \
    curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg && \
    echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_$NODE_MAJOR_VERSION.x nodistro main" | tee /etc/apt/sources.list.d/nodesource.list && \
    apt-get -qq update && apt-get -qq install nodejs -y

# This is where the base image we are using sets the $GOPATH
ENV PATH=/root/go/bin:${PATH}

# Install Wails
ARG WAILS_VERSION=v2.6.0
RUN go install github.com/wailsapp/wails/v2/cmd/wails@${WAILS_VERSION}

ENTRYPOINT [ "/bin/bash" ]