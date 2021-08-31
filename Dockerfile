FROM golang:1.16-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0
COPY . .

RUN go build -o /app/polyhedron ./server
CMD ["/app/polyhedron"]
