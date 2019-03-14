FROM golang:1.12 as golang
ADD . /src
RUN set -x && \
    cd /src && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o goapp

FROM alpine
RUN apk update && apk add ca-certificates

RUN addgroup -g 2000 apprunner
RUN adduser -u 2000 -G apprunner -S apprunner

COPY --from=golang /src/goapp /app/goapp
COPY templates /app/templates
COPY static /app/static

WORKDIR /app
RUN chown -R apprunner:apprunner /app

EXPOSE 8080

USER apprunner

CMD ["/app/goapp"]
