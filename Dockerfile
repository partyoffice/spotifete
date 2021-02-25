FROM golang:1.16-alpine3.13

WORKDIR /go/spotifete

COPY ./ /go/spotifete

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 8410

CMD ["spotifete"]
