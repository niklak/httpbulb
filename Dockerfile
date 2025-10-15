FROM golang:1.25-alpine AS build


ENV APP_ROOT=/httpbulb
ENV APP_NAME=bulb_server

COPY . ${APP_ROOT}
WORKDIR ${APP_ROOT}/cmd/bulb

RUN go build -o ${APP_NAME}


FROM alpine:3.22

RUN apk add --no-cache \
	ca-certificates 

RUN adduser -D httpbulb

COPY --chown=httpbulb:httpbulb --from=build /httpbulb/cmd/bulb/bulb_server /usr/local/bin/bulb_server

USER httpbulb

CMD ["bulb_server"]
