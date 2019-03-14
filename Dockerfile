FROM golang:1.12-alpine as golang
RUN apk add --update gcc musl-dev
ADD . /src
RUN set -x && \
    cd /src && \
    go build -a -o goapp

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
