FROM --platform=$BUILDPLATFORM golang:1.26.2 AS build

ARG TARGETOS
ARG TARGETARCH

ENV APP_NAME=httpbulb
ENV GOARCH=$TARGETARCH
ENV GOOS=$TARGETOS

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o ${APP_NAME} ./cmd/bulb/main.go

FROM gcr.io/distroless/static-debian13
ENV APP_NAME=httpbulb

COPY  --from=build /app/${APP_NAME} /usr/local/bin/${APP_NAME}

USER 65532:65532

ENTRYPOINT ["httpbulb"]
