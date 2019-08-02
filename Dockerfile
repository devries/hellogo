FROM golang:1.12 as golang
ADD go.mod /src/go.mod
ADD go.sum /src/go.sum
RUN set -x && \
  cd /src && \
  go get
ADD . /src
RUN set -x && \
    cd /src && \
    CGO_ENABLED=0 GOOS=linux go build -a -v -o goapp

FROM alpine
RUN apk add --no-cache ca-certificates

RUN addgroup -g 2000 apprunner
RUN adduser -u 2000 -G apprunner -S apprunner

COPY --chown=apprunner:apprunner --from=golang /src/goapp /app/goapp
COPY --chown=apprunner:apprunner templates /app/templates
COPY --chown=apprunner:apprunner static /app/static

WORKDIR /app

USER apprunner

CMD ["/app/goapp"]
