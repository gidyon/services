FROM alpine
LABEL maintainer="gideonhacer@gmail.com"
RUN apk update && \
   apk add ca-certificates && \
   update-ca-certificates && \
   rm -rf /var/cache/apk/* && \
   apk add libc6-compat
WORKDIR /app
COPY account .
EXPOSE 80 443
ENTRYPOINT [ "/app/account", "--config-file", "/app/configs/account.prod.yml" ]