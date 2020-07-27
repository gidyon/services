FROM alpine
LABEL maintainer="gideonhacer@gmail.com"
RUN apk update && \
   apk add ca-certificates && \
   update-ca-certificates && \
   rm -rf /var/cache/apk/* && \
   apk add libc6-compat
WORKDIR /app
COPY employment .
EXPOSE 443 80
ENTRYPOINT [ "/app/employment", "--config-file", "configs/employment.prod.yml" ]