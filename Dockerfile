FROM golang:1.12-alpine as golang
RUN apk add --update gcc musl-dev git
ADD . /src
RUN set -x && \
    cd /src && \
    go build -a -o goapp

FROM alpine
RUN apk update && apk add ca-certificates

RUN addgroup -g 2000 apprunner
RUN adduser -u 2000 -G apprunner -S apprunner

COPY --chown=apprunner:apprunner --from=golang /src/goapp /app/goapp
COPY --chown=apprunner:apprunner templates /app/templates
COPY --chown=apprunner:apprunner static /app/static

WORKDIR /app

USER apprunner

CMD ["/app/goapp"]
