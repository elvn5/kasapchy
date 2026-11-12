# syntax=docker/dockerfile:1
# Общий production-Dockerfile для всех Go-микросервисов.
# Контекст сборки — каталог services/, имя сервиса передаётся через SERVICE:
#   docker build -f services/Dockerfile --build-arg SERVICE=users services
ARG GO_VERSION=1.27

FROM golang:${GO_VERSION}-alpine AS build
ARG SERVICE
RUN test -n "$SERVICE" || (echo "SERVICE build-arg is required" && false)
# Сервис собирается как отдельный модуль (replace ../pkg), go.work не нужен.
ENV CGO_ENABLED=0 GOWORK=off
WORKDIR /src

COPY pkg/go.mod pkg/go.sum ./pkg/
COPY ${SERVICE}/go.mod ${SERVICE}/go.sum ./${SERVICE}/
RUN --mount=type=cache,target=/go/pkg/mod cd ${SERVICE} && go mod download

COPY pkg ./pkg
COPY ${SERVICE} ./${SERVICE}
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd ${SERVICE} && go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${SERVICE}

FROM gcr.io/distroless/static-debian12:nonroot
ENV APP_ENV=production
COPY --from=build /out/app /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]
