package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	uuid "github.com/satori/go.uuid"
)

// Snap is the parsed body from frontend's HTTP POST request
type Snap struct {
	Image string
}

// Result is the outcome of the operation
type Result struct {
	Message string
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

// storeNoteImage stores snap image in S3
func storeNoteImage(key, img string) (arn.ARN, error) {
	imgbucket := os.Getenv("NOTELESS_IMAGE_BUCKET")
	uploader := s3manager.NewUploader(session.New())
	loc := fmt.Sprintf("raw/%v.png", key)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(imgbucket),
		Key:    aws.String(loc),
		Body:   strings.NewReader(img),
	})
	return arn.ARN{Partition: "aws", Service: "s3", Resource: imgbucket + "/" + loc}, err
}

// storeNoteDetections stores Rekognition results (detections) in DynamoDB
func storeNoteDetections(accountID, region, key string, data *rekognition.DetectTextOutput) (arn.ARN, error) {
	notelessTable := os.Getenv("NOTELESS_DETECTIONS_TABLE")
	type Record struct {
		SnapID     string `json:"snapid"`
		Detections *rekognition.DetectTextOutput
	}
	av, err := dynamodbattribute.MarshalMap(Record{
		SnapID:     key,
		Detections: data,
	})
	if err != nil {
		return arn.ARN{}, err
	}
	svc := dynamodb.New(session.New())
	res, err := svc.PutItem(&dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(notelessTable),
	})
	log.Printf("%v", res)
	return arn.ARN{Partition: "aws", Service: "dynamodb", Region: region, AccountID: accountID, Resource: notelessTable + "/" + key}, err
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// extract Account ID from context, needed for DynamoDB result, later:
	lc, _ := lambdacontext.FromContext(ctx)
	larn, err := arn.Parse(lc.InvokedFunctionArn)
	if err != nil {
		return serverError(fmt.Errorf("Can't determine account ID: %v", err))
	}
	// 0. decode the base64 encoded HTTP body, we're expecting the image data there:
	snap := Snap{}
	err = json.Unmarshal([]byte(request.Body), &snap)
	if err != nil {
		return serverError(fmt.Errorf("Can't parse %v as a snap: %v", request.Body, err))
	}
	decodedSnapImage, err := base64.StdEncoding.DecodeString(snap.Image)
	if err != nil {
		return serverError(fmt.Errorf("Can't decode base64 string of snap: %v", err))
	}
	log.Printf("Received image of size %v bytes in good order\n", len(decodedSnapImage))

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

	intakeres := Result{}
	numDetections := len(result.TextDetections)
	switch {
	case numDetections > 0: // we have detections, store images and detections
		log.Printf("Got %v results from Rekognition", numDetections)
		// generate unique note ID (nUUID for short):
		nUUID, err := uuid.NewV4()
		if err != nil {
			return serverError(err)
		}
		// write note image data (in PNG format) to S3 under `raw/$nUUID.png`
		loc, err := storeNoteImage(nUUID.String(), string(decodedSnapImage))
		if err != nil {
			return serverError(err)
		}
		log.Printf("Stored notes image in %v", loc.String())
		// insert detections as JSON blob into DynamoDB table with snapID = $nUUID
		loc, err = storeNoteDetections(larn.AccountID, larn.Region, nUUID.String(), result)
		if err != nil {
			return serverError(err)
		}
		log.Printf("Stored detections in %v", loc.String())
		// generate link with number of raw results, pointing to ../notes/$nUUID
		intakeres.Message = fmt.Sprintf("Found %v fragments, see <a href=\"../notes/\">note %v</a>.",
			numDetections, nUUID.String())
	default: // we haven't detected anything, confirm intake and no note created
		intakeres.Message = "In the snap provided, we were not able to detect text and hence didn't create a note."
	}
	irjson, err := json.Marshal(intakeres)
	if err != nil {
		return serverError(err)
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
			"Content-Type":                "application/json",
		},
		Body: string(irjson),
	}, nil
}

func main() {
	lambda.Start(handler)
}
