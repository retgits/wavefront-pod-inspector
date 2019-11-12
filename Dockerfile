FROM golang:alpine as builder

RUN apk add --no-cache git

RUN mkdir -p /home/app
WORKDIR /home/app
COPY . .

RUN GOPROXY="https://gocenter.io" CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w" -o wavefront .

FROM bitnami/kubectl
COPY --from=builder /home/app/wavefront /bin
COPY --from=builder /home/app/entrypoint.sh /bin
ENTRYPOINT [ "/bin/entrypoint.sh" ]