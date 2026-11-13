FROM golang:1.26-alpine AS build
WORKDIR /src
ENV GOWORK=off CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /marketplace ./cmd/marketplace && go build -trimpath -ldflags="-s -w" -o /moderate ./cmd/moderate
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /marketplace /marketplace
COPY --from=build /moderate /moderate
USER nonroot:nonroot
ENV APP_ENV=production
EXPOSE 8080
ENTRYPOINT ["/marketplace"]
