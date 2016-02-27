FROM golang:onbuild
VOLUME /data
ENTRYPOINT app -db /data/yasuc.db -port 8080
EXPOSE 8080
