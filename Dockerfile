FROM golang:1.13 as builder
WORKDIR /go/src/locker
COPY .  .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /go/bin/locker .

FROM scratch
COPY --from=builder /go/bin/locker .
EXPOSE 80
CMD ["./locker"]