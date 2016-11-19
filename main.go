package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var currentFeedNumber = 0

type NewAlbumFeed struct {
	ID     int
	SQSUrl string
	S3Url  string
}

func returnNewAlbumStreamID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Requested /new")

	nvf := NewAlbumFeed{
		ID:     currentFeedNumber,
		SQSUrl: "You will get the SQS Url here",
		S3Url:  "You will get the S3 url here",
	}

	json.NewEncoder(w).Encode(nvf)
	currentFeedNumber++
}

func postNewPhoto(w http.ResponseWriter, r *http.Request) {
	sess, err := session.NewSession(aws.NewConfig().WithRegion("eu-west-1"))
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}

	svc := sqs.New(sess)

	params := &sqs.SendMessageInput{
		MessageBody:  aws.String("This is a test message"),
		QueueUrl:     aws.String("https://sqs.eu-west-1.amazonaws.com/344405436480/event-emotion-queue"),
		DelaySeconds: aws.Int64(1),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"Key": {
				DataType:    aws.String("String"),
				StringValue: aws.String("String"),
			},
		},
	}
	resp, err := svc.SendMessage(params)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(resp)
}

type PhotoStruct struct {
	Id string `json:"id"`
}

func analyzePhoto(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	var result = new(PhotoStruct)
	err := json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result.Id)

	go func() {
		fmt.Println("goroutine")
		//emotion, err := NewEmotionHandler(os.Getenv("MSFTCOG"))
		//result, err := emotion.Recognize("https://portalstoragewuprod.azureedge.net/emotion/recognition1.jpg")
		//if err != nil {
		//	log.Fatal(err)
		//	return
		//}
		//fmt.Println(result)
	}()
}

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	router := mux.NewRouter()

	router.HandleFunc("/new", returnNewAlbumStreamID).Methods("GET")
	router.HandleFunc("/user", postNewPhoto).Methods("POST")
	router.HandleFunc("/analyze", analyzePhoto).Methods("POST")

	http.Handle("/", router)

	fmt.Printf("Go server started on port %v\n", port)

	err := http.ListenAndServe(":"+port, nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
