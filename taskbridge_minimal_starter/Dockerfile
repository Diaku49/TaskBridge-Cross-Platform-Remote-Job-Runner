FROM golang:1.22-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/taskbridge-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/taskbridge-agent ./cmd/agent

FROM alpine:3.20

COPY --from=build /out/taskbridge-server /usr/local/bin/taskbridge-server
COPY --from=build /out/taskbridge-agent /usr/local/bin/taskbridge-agent

EXPOSE 8080

CMD ["taskbridge-server", "--addr", ":8080"]
