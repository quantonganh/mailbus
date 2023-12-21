FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY mailbus .
EXPOSE 8080
ENTRYPOINT [ "./mailbus" ]