FROM alpine

ADD config.toml /config.toml
ADD https://github.com/teeratpitakrat/hora/releases/download/v0.0.2/hora /hora
#ADD hora /hora
RUN chmod 700 /hora

EXPOSE 8080
CMD ["/hora"]
