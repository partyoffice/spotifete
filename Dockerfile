FROM golang:1.17-alpine AS builder

WORKDIR /go/src/github.com/partyoffice/spotifete
COPY . .

ENV CGO_ENABLED=0
RUN go build -v -trimpath -o ./ ./...

FROM alpine:latest

WORKDIR /opt/spotifete

COPY --from=builder /go/src/github.com/partyoffice/spotifete ./
RUN chmod +x /opt/spotifete/spotifete

EXPOSE 8410
CMD ["/opt/spotifete/spotifete"]
