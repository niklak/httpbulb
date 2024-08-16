FROM golang:1.23-bookworm as build


ENV APP_ROOT=/httpbulb
ENV APP_NAME=bulb_server

COPY . ${APP_ROOT}
WORKDIR ${APP_ROOT}/cmd/bulb

RUN go build -o ${APP_NAME}


FROM golang:1.23-bookworm

RUN useradd -s /bin/bash httpbulb

COPY --chown=httpbulb:httpbulb --from=build /httpbulb/cmd/bulb/bulb_server /usr/local/bin/bulb_server

USER httpbulb

CMD ["bulb_server"]
