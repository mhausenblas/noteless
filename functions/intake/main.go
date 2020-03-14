package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
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

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// 0. decode the base64 encoded HTTP body, we're expecting the image data there:
	snap := Snap{}
	err := json.Unmarshal([]byte(request.Body), &snap)
	if err != nil {
		return serverError(fmt.Errorf("Can't parse %v as a snap: %v", request.Body, err))
	}
	decodedSnapImage, err := base64.StdEncoding.DecodeString(snap.Image)
	if err != nil {
		return serverError(fmt.Errorf("Can't decode base64 string of snap: %v", err))
	}
	fmt.Printf("Received image of size %v bytes in good order\n", len(decodedSnapImage))

	// 1. extract text via Rekognition:
	svc := rekognition.New(session.New())
	result, err := svc.DetectText(&rekognition.DetectTextInput{
		Image: &rekognition.Image{
			Bytes: decodedSnapImage,
		},
	})
	if err != nil {
		return serverError(fmt.Errorf("Can't rekognize: %v", err))
	}
	// fmt.Printf("Rekognition results:\n%v\n", result)
	output, err := json.Marshal(result.TextDetections)
	if err != nil {
		return serverError(fmt.Errorf("Can't encode results: %v", err))
	}

	// TBD: 2. store image as PNG in S3 bucket

	// TBD: 3. insert snap in DynamoDB table (with ref to S3 bucket)

	// Done, send confirmation
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
			"Content-Type":                "application/json",
		},
		Body: string(output),
	}, nil
}

func main() {
	lambda.Start(handler)
}
