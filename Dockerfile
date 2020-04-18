FROM golang AS build
WORKDIR /go/src/github.com/jhunt/up-the-ante

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build ./cmd/tabled \
 && mv tabled /

FROM ubuntu:18.04
COPY --from=build /tabled /usr/bin/tabled
ENTRYPOINT ["tabled"]
