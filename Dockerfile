FROM golang:1.7
MAINTAINER Hugo González Labrador

ADD . /go/src/github.com/clawio/clawiod
WORKDIR /go/src/github.com/clawio/clawiod

RUN go get ./...
RUN go install
RUN mkdir /etc/clawiod/

# Create default config file
RUN cp etc/* > /etc/clawiod/

CMD /go/bin/clawiod -conf /etc/clawiod/monolithic.conf

EXPOSE 1560
