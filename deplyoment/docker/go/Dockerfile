# transient image
FROM golang:1.12.7-alpine3.10 as build
WORKDIR /go/src/app
COPY myapp .
ENV GO111MODULE on
RUN go build -v -o /app
# final image
FROM alpine:3.10
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=build /app /app
CMD ["/app"]