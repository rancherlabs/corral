ARG OS=linux
ARG ARCH=amd64
ARG VERSION=dev

FROM golang:1.18-alpine as mage
ARG OS
ARG ARCH
ARG VERSION
WORKDIR /app

COPY . .

ARG TARGET=build

RUN apk update && \
    apk upgrade && \
    apk add curl git

RUN curl -sLf https://github.com/magefile/mage/releases/download/v1.13.0/mage_1.13.0_Linux-64bit.tar.gz | tar xvzf - -C /usr/bin && chmod +x /usr/bin/mage

ENV GOOS=${OS}
ENV GOARCH=${ARCH}
ENV VERSION=${VERSION}

RUN mage -v ${TARGET} ${VERSION}

FROM scratch
ARG OS
ARG ARCH
WORKDIR /app

COPY --from=mage /app/dist/corral-${OS}-${ARCH} /usr/bin/corral

ENTRYPOINT ["corral"]
