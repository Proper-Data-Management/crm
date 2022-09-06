FROM golang:1.15 AS builder

RUN apt-get update \
 && apt-get install -y  pkg-config libxml2-dev zip libreoffice build-essential \
 && apt-get autoremove -y

WORKDIR /tmp/gobuild

COPY ./golang/go.mod .
COPY ./golang/go.sum .

RUN go mod download

COPY ./golang .


#Test stage
RUN GO111MODULE=on CGO_ENABLED=1 GOOS=linux GOPATH=/gobuild

#Build stage
RUN go build  -a -installsuffix cgo -o ./out/app .

FROM ubuntu:18.04

RUN apt-get update \
 && apt-get install -y libxml2-dev zip libreoffice \
 && apt-get autoremove -y \
 && mkdir -p /opt/crm

COPY ./test/static /opt/crm/static
COPY ./test/uploads /opt/crm/uploads
COPY  --from=builder /tmp/gobuild/out/app /opt/crm/app

WORKDIR /opt/crm

RUN chmod +x ./app

CMD ["./app"]
