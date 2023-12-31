FROM alpine:latest as alpine
RUN apk add -U --no-cache ca-certificates

FROM scratch
LABEL maintainer=ysh9579
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENV BD_CONFIG /root/.buddle/api.config.yaml

COPY bin/go-api /usr/bin/api

ENTRYPOINT ["/usr/bin/api"]