FROM alpine:3.14

WORKDIR /opt/spotifete

COPY ./spotifete /opt/spotifete/spotifete
COPY ./resources /opt/spotifete/resources

RUN chmod +x /opt/spotifete/spotifete

EXPOSE 8410

CMD ["/opt/spotifete/spotifete"]
