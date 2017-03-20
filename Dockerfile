FROM golang:1-alpine

RUN apk add --no-cache curl bash docker

ADD ["main.go", "deploy.sh", "/go/src/server/"]
WORKDIR /go/src/server/

ENV CGO_ENABLED=0
RUN go build -a --installsuffix cgo -ldflags="-s -w"

EXPOSE 5000
CMD ["./server"]
