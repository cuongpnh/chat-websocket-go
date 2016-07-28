FROM golang:1.6.2-alpine
MAINTAINER Cuong Pham <phamnguyenhungcuong@gmail.com>
RUN apk update && apk upgrade && apk add --no-cache git openssh
ENV APP_HOME $GOPATH/src/tracker
RUN mkdir -p $APP_HOME
WORKDIR $APP_HOME
RUN go get github.com/codegangsta/gin
RUN go get github.com/tools/godep
COPY ./Godeps $APP_HOME/Godeps
RUN godep restore
COPY . $APP_HOME
RUN chmod 775 $GOPATH/src/tracker/docker-start.sh
CMD ["/bin/sh", "./docker-start.sh"]
