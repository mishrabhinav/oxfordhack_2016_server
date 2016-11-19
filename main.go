package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

var currentFeedNumber = 0

type NewVideoFeed struct {
	ID     int
	SQSUrl string
	S3Url  string
}

func returnNewVideoStreamID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Requested /new")

	nvf := NewVideoFeed{
		ID:     currentFeedNumber,
		SQSUrl: "You will get the SQS Url here",
		S3Url:  "You will get the S3 url here",
	}

	json.NewEncoder(w).Encode(nvf)
	currentFeedNumber++
}

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	http.HandleFunc("/new", returnNewVideoStreamID) // set router

	fmt.Printf("Go server started on port %v\n", port)

	err := http.ListenAndServe(":"+port, nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
