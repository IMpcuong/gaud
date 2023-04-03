FROM golang:1.18-alpine as build
LABEL maintainer="IMpcuong"
ARG APP=gad

WORKDIR /usr/src/app

ENV GO111MODULE=on \
    CGO_ENABLED=0  \
    GOARCH="amd64" \
    GOOS=linux

COPY go.mod ./
RUN if [[ -f go.sum ]]; then cp go.sum ./; fi
RUN go clean && \
  go mod tidy && \
  go mod download && \
  go mod verify

COPY *.go ./
RUN go build --ldflags "-extldflags -static" \
      -v -o ./$APP ./...

# RUN apk add --update --no-cache
FROM golang:1.18-alpine as main

WORKDIR /usr/local/bin/app

COPY --from=build /usr/src/app/gad .
# EXPOSE 80/tcp 443/tcp

# Example download through docker-image:
ENTRYPOINT [ "./gad" ]
CMD [ "-d" ]
