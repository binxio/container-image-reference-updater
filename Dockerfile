FROM golang:1.12

WORKDIR /app
ADD	. /app
RUN go mod download
RUN	go build -o container-image-reference-updater -ldflags '-extldflags "-static"' main.go

FROM gcr.io/cloud-builders/gcloud:latest

ENV REPOSITORY_PROJECT=my-serverless-app-cicd
ENV REPOSITORY_NAME=infrastructure
ENV GIT_USER_NAME="Container image reference updater"
ENV PATH=$PATH:/usr/local/bin

ADD update-image-references /usr/local/bin/
COPY --from=0 /app/container-image-reference-updater /usr/local/bin/

ENTRYPOINT 	["/usr/local/bin/container-image-reference-updater"]
EXPOSE 		8080
