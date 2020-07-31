FROM golang:1.12
RUN mkdir /dns-stress
ADD main.go /dns-stress
ADD domain /dns-stress
WORKDIR /dns-stress
RUN go build
ENTRYPOINT  ["./dns-stress"]