FROM golang:1.18 AS build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o spanet_exporter .

FROM alpine:3.15

RUN apk --no-cache add ca-certificates
WORKDIR /app/
COPY --from=build /build/spanet_exporter .
ENTRYPOINT ["/app/spanet_exporter"]
