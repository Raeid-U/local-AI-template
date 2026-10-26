FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY server ./server
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/local-ai ./server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/local-ai /local-ai
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/local-ai"]
