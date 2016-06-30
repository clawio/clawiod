FROM golang:1.6
MAINTAINER Hugo González Labrador

ADD . /go/src/github.com/clawio/clawiod
WORKDIR /go/src/github.com/clawio/clawiod

RUN go get ./...
RUN go install
RUN mkdir /etc/clawiod/

# Create default config file
RUN echo "{}" > /etc/clawiod/clawiod.conf

CMD /go/bin/clawiod -config /etc/clawiod/clawiod.conf

EXPOSE 1502
