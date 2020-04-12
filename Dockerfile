FROM phev8/go-dep-builder:latest as builder
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates
RUN mkdir -p /go/src/github.com/influenzanet/study-service
ENV GO111MODULE=on
ADD . /go/src/github.com/influenzanet/study-service/
WORKDIR /go/src/github.com/influenzanet/study-service
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o app .
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/influenzanet/study-service/app /app/
WORKDIR /app
EXPOSE 5200:5200
CMD ["./app"]
