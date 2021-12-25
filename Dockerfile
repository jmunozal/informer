# syntax=docker/dockerfile:1
FROM golang:1.16 AS build

WORKDIR "/app"
COPY src/main.go ./
COPY src/go.mod ./

RUN go mod download
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /informer

# STAGE 2: Deployment
#FROM alpine:latest
FROM scratch

WORKDIR "/"

EXPOSE 8080

#USER nobody:nobody
COPY --from=build /informer /informer

CMD [ "/informer" ]