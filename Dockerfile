FROM golang:1.21 AS builder
COPY . /build
WORKDIR /build
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o deltans .


FROM alpine:3.19
COPY --from=builder /build/deltans /deltans
EXPOSE 5333/udp
ENTRYPOINT /deltans
