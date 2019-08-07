FROM golang:alpine as builder

RUN apk add --no-cache git

RUN mkdir -p /home/app
WORKDIR /home/app
COPY . .

RUN GOPROXY="https://gocenter.io" CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w" -o wavefront .

FROM alpine:3.10
COPY --from=builder /home/app/wavefront /bin
RUN apk add --no-cache ca-certificates
CMD [ "wavefront" ]