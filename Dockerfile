FROM golang:1.18-alpine as build
LABEL maintainer="IMpcuong"
ARG APP=gad

WORKDIR /usr/src/app

COPY go.mod ./
RUN if [[ -f go.sum ]]; then cp go.sum ./; fi
RUN go clean && \
  go mod tidy && \
  go mod download && \
  go mod verify

COPY *.go ./
RUN go build -v -o ./$APP ./...

# RUN apk add --update --no-cache
FROM golang:1.18-alpine as main

WORKDIR /usr/local/bin/app

COPY --from=build /usr/src/app/gad .
# EXPOSE 80/tcp 443/tcp

# Example download through docker-image:
ENTRYPOINT [ "./gad" ]
CMD [ "-d" ]