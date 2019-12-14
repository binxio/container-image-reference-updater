FROM golang:1.12

WORKDIR /app
ADD	. /app
RUN go mod download
RUN	go build -o gcr-docker-image-updater -ldflags '-extldflags "-static"' main.go

FROM gcr.io/cloud-builders/gcloud:latest

ENV REPOSITORY_PROJECT=my-serverless-app-cicd
ENV REPOSITORY_NAME=infrastructure
ENV PATH=$PATH:/usr/local/bin

ADD update-image-reference /usr/local/bin/
COPY --from=0 /app/gcr-docker-image-updater /usr/local/bin/

ENTRYPOINT 	["/usr/local/bin/gcr-docker-image-updater"]
EXPOSE 		8080
