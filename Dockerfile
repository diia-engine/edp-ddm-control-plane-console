FROM golang:1.17.9-alpine

RUN apk update && apk add ca-certificates git openssh-client

ENV USER_UID=1001 \
    USER_NAME=admin-console \
    HOME=/home/admin-console

RUN addgroup --gid ${USER_UID} ${USER_NAME} \
    && adduser --disabled-password --uid ${USER_UID} --ingroup ${USER_NAME} --home ${HOME} ${USER_NAME}

WORKDIR /go/bin
ENV ZONEINFO=/usr/local/go/lib/time/zoneinfo.zip

COPY static static
COPY frontend frontend
COPY templates templates
COPY default.env .
COPY locale locale
COPY osplm.ini .
RUN mkdir /home/admin-console/.ssh && chown ${USER_NAME}:${USER_NAME} /home/admin-console/.ssh && chmod 700 /home/admin-console/.ssh
COPY ssh-config.txt /home/admin-console/.ssh/config
RUN chown ${USER_NAME}:${USER_NAME} /home/admin-console/.ssh/config && chmod 700 /home/admin-console/.ssh/config
USER ${USER_UID}
RUN git config --global user.email "admin@localhost"
RUN git config --global user.name "admin"
COPY control-plane-console .
CMD ["/go/bin/control-plane-console"]