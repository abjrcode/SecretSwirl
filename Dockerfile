ARG GO_VERSION=1.21.1

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} as builder

WORKDIR /usr/src/app

# Install Libgtk, webkit and NSIS
RUN apt -qq update && apt -qq install -y libgtk-3-dev libwebkit2gtk-4.0-dev nsis

ARG NODE_MAJOR_VERSION=18

# Install NodeJS
RUN apt-get -qq install -y ca-certificates curl gnupg && \
    mkdir -p /etc/apt/keyrings && \
    curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg && \
    echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_$NODE_MAJOR_VERSION.x nodistro main" | tee /etc/apt/sources.list.d/nodesource.list && \
    apt update && apt-get -qq install nodejs -y

# Install Wails
ARG WAILS_VERSION=v2.6.0
RUN go install github.com/wailsapp/wails/v2/cmd/wails@${WAILS_VERSION}

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG RELEASE_TAG=v0.0.0-0-g000000
ARG BUILD_TIMESTAMP=NOW
ARG COMMIT_SHA=docker
ARG BUILD_LINK=http://docker.local
RUN go run mage.go build ${RELEASE_TAG} ${BUILD_TIMESTAMP} ${COMMIT_SHA} ${BUILD_LINK}


# I REALLY TRIED - ATTEMPT TO CROSS COMPILE BY USING GCC FOR AARCH64
# IT  DID NOT WORK BECAUSE QEMU (SEEMINGLY USED BY WAILS) JUMPS FROM ONE ERROR TO ANOTHER

# - https://github.com/wailsapp/wails/issues/2833#issuecomment-1685684654
# - https://github.com/wailsapp/wails/issues/2060

#############################################################

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}

COPY --from=builder /usr/src/app/build/bin /out

ENTRYPOINT [ "sh", "-c" ]
CMD [ "cp -r /out/. /artifacts/" ]