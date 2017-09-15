From alpine:3.6
RUN apk update && \
    apk add ca-certificates
COPY riker /
