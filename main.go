package main

import (
	"encoding/json"
	"fmt"
	"github.com/docker/distribution/reference"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

type GCRNotification struct {
	Tag    string `json:"tag"`
	Action string `json:"action""`
	Digest string `json:"digest,omitempty"`
}

type PubSubMessage struct {
	Message struct {
		Data []byte `json:"data,omitempty"`
		ID   string `json:"id"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func HandleContainerRegistryEvent(w http.ResponseWriter, r *http.Request) {
	var m PubSubMessage
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read the request body", http.StatusBadRequest)
		return
	}
	log.Printf("body -> %s", body)
	if err := json.Unmarshal(body, &m); err != nil {
		log.Printf("%v", m)
		http.Error(w, "not a Google Pub/Sub message", http.StatusBadRequest)
		return
	}
	var notification GCRNotification
	if err := json.Unmarshal(m.Message.Data, &notification); err != nil {
		http.Error(w, "not a Google Container Registry notification", http.StatusBadRequest)
		return
	}
	repository, err := reference.Parse(notification.Tag)
	if err != nil {
		http.Error(w, "tag is invalid docker repository reference", http.StatusBadRequest)
		return
	}

	if notification.Action != "INSERT" {
		fmt.Fprintf(w, "ignoring %s action", notification.Action)
		return
	}

	tag, ok := repository.(reference.Tagged)
	if !ok || tag.Tag() == "" {
		fmt.Fprint(w, "ignoring image repository update without tag")
		return
	}
	if tag.Tag() == "latest" {
		fmt.Fprint(w, "ignoring image repository update 'latest' tag")
		return
	}
	err = updateGitRepository(repository.String())
	if err != nil {
		http.Error(w, "failed to update images reference", http.StatusInternalServerError)
	} else {
		fmt.Fprintf(w, "updated image reference %s\n", repository.String())
	}
}

func updateGitRepository(repositoryReference string) error {
	cmd := exec.Command("/usr/local/bin/update-image-references", repositoryReference)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	http.HandleFunc("/", HandleContainerRegistryEvent)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
