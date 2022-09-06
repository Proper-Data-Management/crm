FROM ubuntu:18.04

RUN apt-get update \
 && apt-get install -y libxml2-dev zip \
 && apt-get autoremove -y \
 && mkdir /opt/crms \
 && apt-get install -y  libreoffice

COPY ./static /opt/crm/static
COPY ./uploads /opt/crm/uploads
COPY ./app /opt/crm/app

WORKDIR /opt/crm

CMD ["./app"]
