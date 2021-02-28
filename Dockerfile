FROM golang:3.13

WORKDIR /opt/spotifete

COPY ./spotifete ./spotifete
COPY ./resources ./resources

EXPOSE 8410

CMD ["./spotifete"]
