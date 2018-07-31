FROM golang:1.11-alpine as build-backend

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /go/src/github.com/asnelzin/stream-api
ADD . /go/src/github.com/asnelzin/stream-api

RUN go build -o stream-api -ldflags "-X main.revision=local -s -w" ./cmd/stream-api

FROM asnelzin/baseimage:latest

WORKDIR /srv

ADD entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY --from=build-backend /go/src/github.com/asnelzin/stream-api/stream-api /srv/stream-api
RUN chown -R app:app /srv
RUN ln -s /srv/stream-api /usr/bin/stream-api

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD curl --fail http://localhost:8080/ping || exit 1

ENTRYPOINT ["/entrypoint.sh"]
