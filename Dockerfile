# Build
FROM golang:1.24.4-alpine3.22 AS build

RUN apk add upx alpine-sdk

WORKDIR /build

COPY go.mod go.sum ./ 
RUN go mod download 

COPY . .

RUN CGO_ENABLED=1 go build -ldflags "-s -w" -v -tags musl -o main .
RUN upx --best --lzma main

# End container
FROM alpine:3.22

WORKDIR /immich

COPY --from=build /build/main .

RUN chmod +x ./main

ENTRYPOINT ["./main"]
