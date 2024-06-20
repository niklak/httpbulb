FROM golang:1.22-bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libnss3-tools \
    git\
    && apt-get clean \
    && apt-get autoremove \
    && rm -rf /var/lib/apt/lists/*


RUN useradd -ms /bin/bash httpbulb

ENV HOME=/home/httpbulb
ENV APP_ROOT=${HOME}/httpbulb

COPY . ${APP_ROOT}

WORKDIR ${APP_ROOT}


WORKDIR ${APP_ROOT}/cmd/bulb

RUN go build -o bulb_server

RUN chown -R httpbulb:httpbulb ${APP_ROOT}

USER httpbulb

ENTRYPOINT ["./bulb_server"]
