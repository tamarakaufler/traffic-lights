FROM golang:1.13.1-buster as builder
WORKDIR /app
COPY . .

ENV GOARCH arm
ENV GOARM 7
ENV GOOS linux
RUN ["go", "build", "-o", "trafficlights", "."]

FROM raspbian/jessie:latest
WORKDIR /app
COPY --from=builder /app/trafficlights /app
CMD ["/app/trafficlights"]
