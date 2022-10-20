FROM golang:1.17.9-alpine

RUN apk update && apk add ca-certificates git openssh-client

ENV USER_UID=1001 \
    USER_NAME=admin-console \
    HOME=/home/admin-console

RUN addgroup --gid ${USER_UID} ${USER_NAME} \
    && adduser --disabled-password --uid ${USER_UID} --ingroup ${USER_NAME} --home ${HOME} ${USER_NAME}

WORKDIR /go/bin
ENV ZONEINFO=/usr/local/go/lib/time/zoneinfo.zip

COPY control-plane-console .
COPY static static
COPY templates templates
COPY default.env .
COPY locale locale
COPY osplm.ini .
RUN mkdir /home/admin-console/.ssh && chown ${USER_NAME}:${USER_NAME} /home/admin-console/.ssh && chmod 700 /home/admin-console/.ssh
COPY ssh-config.txt /home/admin-console/.ssh/config
RUN chown ${USER_NAME}:${USER_NAME} /home/admin-console/.ssh/config && chmod 700 /home/admin-console/.ssh/config

USER ${USER_UID}

CMD ["/go/bin/control-plane-console"]