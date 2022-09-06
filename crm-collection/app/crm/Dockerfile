FROM ubuntu:18.04

RUN apt-get update \
 && apt-get install -y libxml2-dev zip \
 && apt-get autoremove -y \
 && mkdir /opt/crm

WORKDIR /opt/crm

CMD ["./app"]
