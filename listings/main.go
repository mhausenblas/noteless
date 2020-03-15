package main

import (
	"context"
	"encoding/json"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/open-policy-agent/opa/rego"
)

var (

	// module is the set of Rego rules to detect commands
	module = `package noteless

# we declare everything a command that is
# 1. at least two characters long, 2. recognized with at least 96% confidence, and
# 3. in our command list
detected_commands[msg] {
	dt := input[0].Detections.TextDetections[_].DetectedText
	confidence := input[0].Detections.TextDetections[_].Confidence
	count(dt) > 1
	confidence > 96.0
	iscommand(dt)
	msg := sprintf("%v", [lower(dt)])
}

# checks if a word is a command
iscommand(candidate) {
	allcommands := ["go", "stop", "on", "off", "left", "right", "up", "down", "to"]
	allcommands[_] = lower(candidate)
}`
)

func main() {
	srv := &http.Server{Addr: ":9898"}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		http.HandleFunc("/rules", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(module))
		})
		http.HandleFunc("/notes", func(w http.ResponseWriter, r *http.Request) {
			ni, err := notesIcons()
			if err != nil {
				http.Error(w, err.Error(), 500)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(ni)
		})
		http.HandleFunc("/commands", func(w http.ResponseWriter, r *http.Request) {
			dt, err := detectedTexts()
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			dc, err := commands(dt)
			if err != nil {
				http.Error(w, err.Error(), 500)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(dc)
		})
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Print("Server Started")
	<-done
	log.Print("Server Stopped")
}

// detectedTexts queries the DDB table for detected texts
func detectedTexts() ([]interface{}, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"),
	})
	if err != nil {
		return nil, err
	}
	svc := dynamodb.New(sess)
	result, err := svc.Scan(&dynamodb.ScanInput{
		TableName: aws.String("Noteless"),
	})
	var notes []interface{}
	for i := range result.Items {
		var r interface{}
		err = dynamodbattribute.UnmarshalMap(result.Items[i], &r)
		if err != nil {
			return nil, err
		}
		notes = append(notes, r)
	}
	return notes, nil
}

// commands returns a list of detected commands
func commands(input []interface{}) ([]byte, error) {
	reg := rego.New(
		rego.Query("data.noteless.detected_commands"),
		rego.Module("commands.rego", module),
		rego.Input(input),
	)
	ctx := context.Background()
	rs, err := reg.Eval(ctx)
	if err != nil {
		return []byte(""), err
	}
	val, err := json.Marshal(rs[0].Expressions[0].Value)
	if err != nil {
		return []byte(""), err
	}
	return val, nil
}

func notesIcons() ([]byte, error) {
	type NoteIcon struct {
		Content string
	}
	ni := NoteIcon{Content: "PNG IMAGE 0"}
	val, err := json.Marshal(ni)
	if err != nil {
		return []byte(""), err
	}
	return val, nil
}
