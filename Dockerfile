# ----------------------------------------------------
# Base
# ----------------------------------------------------
FROM golang:1.15-alpine3.12 AS build_base
RUN mkdir -p {/app/dolittle/app/bin}
WORKDIR /app/dolittle

COPY go.mod .
COPY go.sum .
RUN go mod download

# ----------------------------------------------------
# Build + Test
# ----------------------------------------------------
ARG GIT_COMMIT
ENV GIT_COMMIT ${GIT_COMMIT}
ARG GIT_HASH_DATE
ENV GIT_HASH_DATE ${GIT_HASH_DATE}

FROM build_base AS build
WORKDIR /app/dolittle
COPY --from=build_base /app/dolittle .
COPY . .

ENV GOOS linux
ENV GOARCH amd64
ENV CGO_ENABLED 0
RUN go build -ldflags "-s -w " -o app main.go

# ----------------------------------------------------
# Release
# ----------------------------------------------------
FROM alpine:3.12 AS release
ENV LC_ALL=en_US.UTF-8
ENV LC_LANG=en_US.UTF-8
ENV LC_LANGUAGE=en_US.UTF-8

RUN mkdir -p {/app/bin}
COPY --from=build /app/dolittle/app /app/bin/app

WORKDIR /app
ENTRYPOINT ["/app/bin/app"]

EXPOSE 8000