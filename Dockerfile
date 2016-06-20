FROM golang:1.6
MAINTAINER Hugo González Labrador

ADD . /go/src/github.com/clawio/clawiod
WORKDIR /go/src/github.com/clawio/clawiod

RUN go get ./...
RUN go install

CMD /go/bin/clawiod

EXPOSE 1502
