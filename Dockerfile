FROM golang:1.22-bookworm as build


ENV APP_ROOT=/httpbulb
ENV APP_NAME=bulb_server

COPY . ${APP_ROOT}
WORKDIR ${APP_ROOT}/cmd/bulb

RUN go build -o ${APP_NAME}


FROM golang:1.22-bookworm


RUN useradd -s /bin/bash httpbulb

ENV APP_NAME=bulb_server

COPY --chown=httpbulb:httpbulb --from=build /httpbulb/cmd/bulb/${APP_NAME} /usr/local/bin/${APP_NAME}

USER httpbulb

ENTRYPOINT ${APP_NAME}
