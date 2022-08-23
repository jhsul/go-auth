#FROM ubuntu:trusty
#RUN sudo apt-get -y update
#RUN sudo apt-get -y upgrade
#RUN sudo apt-get install -y sqlite3 libsqlite3-dev


#RUN mkdir /db
#RUN /usr/bin/sqlite3 ./database.db
#CMD /bin/bash

FROM golang:1.19.0-bullseye AS builder

#RUN apk add --no-cache git

#ENV GO111MODULE=on
#ENV GOFLAGS=-mod=vendor

WORKDIR /src

COPY . .
#COPY go.mod .
#COPY go.sum .

RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -o /app -a -ldflags '-linkmode external -extldflags "-static"' .

#RUN go build -o /app .

FROM scratch
COPY --from=builder /app /app
#COPY --from=builder /src/build /build


EXPOSE 3000

# Start fresh from a smaller image
#FROM alpine:latest
#RUN apk add ca-certificates

#COPY --from=build_base /tmp/go-auth/out/go-auth /app/go-auth


ENTRYPOINT [ "/app" ]