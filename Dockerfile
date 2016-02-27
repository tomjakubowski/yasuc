FROM golang:onbuild
ENTRYPOINT app -db /data/yasuc.db -port 8080
EXPOSE 8080
