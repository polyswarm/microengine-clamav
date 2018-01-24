FROM golang:latest
LABEL maintainer="Maxwell Koo <mjkoo90@gmail.com>"

ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

RUN mkdir -p $GOPATH/src/clamav-microengine/
ADD . $GOPATH/src/clamav-microengine/

#RUN set -x && \
#    cd $GOPATH/src/clamav-microengine && \
#    go get . && \
#    go install
ADD clamav-microengine $GOPATH/bin/clamav-microengine

RUN useradd -ms /bin/bash microengine
WORKDIR /home/microengine
USER microengine

ENTRYPOINT $GOPATH/bin/clamav-microengine
