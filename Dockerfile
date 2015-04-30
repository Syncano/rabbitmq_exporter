FROM golang
MAINTAINER Michal Kobus <michal.kobus@syncano.com>
ENTRYPOINT [ "go-wrapper", "run" ]
EXPOSE 9672
