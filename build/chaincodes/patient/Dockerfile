FROM alpine
LABEL maintainer="gideonhacer@gmail.com"
RUN apk update && \
   apk add ca-certificates && \
   update-ca-certificates && \
   rm -rf /var/cache/apk/* && \
   apk add libc6-compat
WORKDIR /app
EXPOSE 443 80
COPY patient .
ENTRYPOINT [ "/app/patient", "--config-file", "configs/patient.prod.yml" ]