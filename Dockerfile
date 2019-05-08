FROM golang:1.12-alpine3.9 as builder
WORKDIR /workspace
ADD . /workspace
RUN go build -o app

FROM alpine:3.9
WORKDIR /app
COPY --from=builder /workspace/app /app/app
ENTRYPOINT ./app
