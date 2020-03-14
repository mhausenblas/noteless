package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Snap is the parsed body from frontend's HTTP POST request
type Snap struct {
	Image string
}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
	fmt.Println(err.Error())
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
		Body: fmt.Sprintf("%v", err.Error()),
	}, nil
}

// func detectText() {
// 	sess := session.New(&aws.Config{
// 		Region: aws.String("eu-west-1"),
// 	})
// 	svc := rekognition.New(sess)
// 	if r.Body == nil {
// 		http.Error(w, "Request Body Expected", 400)
// 		return
// 	}
// 	var parsed Snap
// 	err := json.NewDecoder(r.Body).Decode(&parsed)
// 	if err != nil {
// 		http.Error(w, err.Error(), 400)
// 	}
// 	decodedImage, err := base64.StdEncoding.DecodeString(parsed.Image)
// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 		return
// 	}
// 	input := &rekognition.DetectTextInput{
// 		Image: &rekognition.Image{
// 			Bytes: decodedImage,
// 		},
// 	}
// 	result, err := svc.DetectText(input)
// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 		return
// 	}
// 	output, err := json.Marshal(result)
// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Header().Set("Access-Control-Allow-Origin", "*")
// 	w.WriteHeader(http.StatusOK)
// 	w.Write(output)

// }

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	snap := Snap{}
	err := json.Unmarshal([]byte(request.Body), &snap)
	if err != nil {
		return serverError(fmt.Errorf("Can't parse %v as a snap: %v", request.Body, err))
	}
	decodedSnapImage, err := base64.StdEncoding.DecodeString(snap.Image)
	if err != nil {
		return serverError(fmt.Errorf("Can't decode base64 string of snap: %v", err))
	}
	fmt.Printf("%+v\n", snap)
	result := len(decodedSnapImage)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
			"Content-Type":                "text/plain",
		},
		Body: fmt.Sprintf("%v", result),
	}, nil
}

func main() {
	lambda.Start(handler)
}
