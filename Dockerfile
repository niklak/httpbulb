FROM --platform=$BUILDPLATFORM golang:1.26.2 AS build


ARG TARGETOS
ARG TARGETARCH

ENV GOARCH=$TARGETARCH
ENV GOOS=$TARGETOS

ENV APP_ROOT=/httpbulb
ENV APP_NAME=httpbulb
ENV SERVER_HOST=0.0.0.0

WORKDIR ${APP_ROOT}

COPY go.mod go.sum ./
RUN go mod download

COPY . .
WORKDIR ${APP_ROOT}/cmd/bulb

RUN CGO_ENABLED=0 go build -o ${APP_NAME}


FROM gcr.io/distroless/static-debian13


COPY  --from=build /httpbulb/cmd/bulb/httpbulb /usr/local/bin/httpbulb

USER 65532:65532

ENTRYPOINT ["httpbulb"]
