package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

const queueUrl = "QUEUE URL"
const apiKey = "API KEY"

func readQueue(w http.ResponseWriter, r *http.Request) {
	sess, err := session.NewSession(aws.NewConfig().WithRegion("eu-west-1"))
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}

	svc := sqs.New(sess)

	params := &sqs.ReceiveMessageInput{
		QueueUrl:     aws.String(queueUrl),
		MaxNumberOfMessages: aws.Int64(1),
		MessageAttributeNames: []*string{
			aws.String(".*"),
		},
	VisibilityTimeout: aws.Int64(1),
	WaitTimeSeconds:   aws.Int64(1),
	}
	resp, err := svc.ReceiveMessage(params)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	emotions := make(chan Score, len(resp.Messages[0].MessageAttributes))

	for _, message := range resp.Messages {
		for _, attribute := range message.MessageAttributes {
			m := *attribute.StringValue

			go func(m string) {
				emotion, err := NewEmotionHandler(apiKey)
				result, err := emotion.Recognize(m)
				if err != nil {
					log.Fatal(err)
					return
				}

				var imageScore Score
				for _, detail := range result {
					imageScore.Anger     += detail.Scores.Anger
					imageScore.Contempt  += detail.Scores.Contempt
					imageScore.Disgust   += detail.Scores.Disgust
					imageScore.Fear      += detail.Scores.Fear
					imageScore.Happiness += detail.Scores.Happiness
					imageScore.Neutral   += detail.Scores.Neutral
					imageScore.Sadness   += detail.Scores.Sadness
					imageScore.Surprise  += detail.Scores.Surprise
				}

				emotions <- imageScore
			}(m)
		}
	}

	defer func() {
		params := &sqs.DeleteMessageInput{
			QueueUrl:     aws.String(queueUrl),
			ReceiptHandle: aws.String(*(resp.Messages[0].ReceiptHandle)),
		}
		_, err := svc.DeleteMessage(params)

		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}()

	defer func() {
		emotion := <-emotions
		js, err := json.Marshal(emotion)

		if err != nil {
			fmt.Println(err)
		}

		result := fmt.Sprintf("[%v,", string(js))
		for i := 0; i < len(resp.Messages[0].MessageAttributes) - 2; i++ {
			emotion = <-emotions
			js, err = json.Marshal(emotion)
			if err != nil {
				fmt.Println(err)
			}
			result = fmt.Sprintf("%v%v,", result, string(js))
		}

		emotion = <-emotions
		js, err = json.Marshal(emotion)
		if err != nil {
			fmt.Println(err)
		}

		result = fmt.Sprintf("%v%v]", result, string(js))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(result))

	}()
}

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		port = "3030"
	}

	router := mux.NewRouter()
	router.HandleFunc("/read", readQueue).Methods("GET")

	http.Handle("/", router)

	fmt.Printf("Go server started on port %v\n", port)

	err := http.ListenAndServe(":"+port, nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
