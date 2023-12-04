# FROM cgr.dev/chainguard/go:latest-dev AS builder
FROM cgr.dev/chainguard/go:latest as builder

LABEL org.opencontainers.image.authors="raffael@nice.pink"
LABEL org.opencontainers.image.source="https://github.com/nice-pink/go-ops"

WORKDIR /app

# get go module ready
COPY ./go.mod ./
RUN go mod download

# copy module code
COPY . .

RUN ./build_all

####################################################################################################

# FROM cgr.dev/chainguard/go:latest AS sitesync-runner
FROM cgr.dev/chainguard/glibc-dynamic:latest AS request

COPY --from=builder /app/bin/request /request
ENTRYPOINT [ "/request" ]

####################################################################################################

FROM cgr.dev/chainguard/glibc-dynamic:latest

COPY --from=builder /app/bin/* .

# ENTRYPOINT [ "/request" ]

