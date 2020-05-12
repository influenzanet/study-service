##########################
# STAGE 1
##########################
FROM golang:1.14-alpine as builder
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates
RUN mkdir -p /go/src/github.com/influenzanet/study-service
ENV GO111MODULE=on
ADD . /go/src/github.com/influenzanet/study-service/
WORKDIR /go/src/github.com/influenzanet/study-service
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
WORKDIR /go/src/github.com/influenzanet/study-service/cmd/study-service-app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o app .

##########################
# STAGE 2
##########################
FROM scratch
# we need those on scratch for connecting to mongoDB:
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/influenzanet/study-service/cmd/study-service-app/app /app/
WORKDIR /app
EXPOSE 5003:5003
CMD ["./app"]
