FROM alpine
LABEL maintainer="gideonhacer@gmail.com"
RUN apk update && \
   apk add ca-certificates && \
   update-ca-certificates && \
   rm -rf /var/cache/apk/*
EXPOSE 80 443 8080 9090
WORKDIR /app
COPY service .
ENTRYPOINT [ "/app/service" ]
CMD ["--config-file", "configs/config.yml"]