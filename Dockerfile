FROM golang:1.7.4
ENV WD $GOPATH/src/github.com/lyft/goruntime
RUN curl https://glide.sh/get | sh

ADD . $WD
WORKDIR $WD
