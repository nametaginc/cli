FROM alpine
RUN apk add --no-cache ca-certificates
ADD nametag /usr/bin/nametag
