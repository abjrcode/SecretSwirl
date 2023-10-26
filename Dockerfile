ARG BASE_IMAGE=swervo_base:latest

FROM --platform=$BUILDPLATFORM ${BASE_IMAGE} as builder

COPY go.mod go.sum ./

RUN go mod download

COPY . .


ARG BUILD_TYPE=debug
ARG RELEASE_TAG=v0.0.0-0-g000000
ARG BUILD_TIMESTAMP=NOW
ARG COMMIT_SHA=docker
ARG BUILD_LINK=http://docker.local

ENV CGO_ENABLED=1

RUN go run mage.go build ${BUILD_TYPE} ${RELEASE_TAG} ${BUILD_TIMESTAMP} ${COMMIT_SHA} ${BUILD_LINK}

RUN go test -v ./...

#############################################################

FROM --platform=$BUILDPLATFORM ${BASE_IMAGE}

COPY --from=builder /usr/src/app/build/bin /out

ENTRYPOINT [ "sh", "-c" ]
CMD [ "cp -r /out/. /artifacts/" ]