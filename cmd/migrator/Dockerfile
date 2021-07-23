FROM alpine
LABEL maintainer="gideonhacer@gmail.com"
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY migrator .
ENTRYPOINT [ "/app/migrator" ]