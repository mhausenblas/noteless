package main

import (
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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	uuid "github.com/satori/go.uuid"
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

// storeClusterSpec stores the cluster spec in a given bucket
func storeNoteImage(key, img string) (arn.ARN, error) {
	uploader := s3manager.NewUploader(session.New())
	imgbucket := os.Getenv("NOTELESS_IMAGE_BUCKET")
	loc := fmt.Sprintf("raw/%v.png", key)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(imgbucket),
		Key:    aws.String(loc),
		Body:   strings.NewReader(img),
	})
	return arn.ARN{Partition: "aws", Service: "s3", Resource: imgbucket + "/" + loc}, err
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

	// 2. if we have results, store image as PNG in S3 bucket
	//    and insert detections in DynamoDB table (with pointer to S3 bucket)
	numDetections := len(result.TextDetections)
	if numDetections > 0 {
		log.Printf("Got %v results from Rekognition", numDetections)
		output, err := json.Marshal(result)
		if err != nil {
			return serverError(fmt.Errorf("Can't encode results: %v", err))
		}
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
		// put detections as JSON blog into DynamoDB table with snapID = $nUUID
		_ = output

		// generate link with number of raw results, pointing to ../notes/$nUUID
		noteLink := fmt.Sprintf("Found %v fragments in snap, see <a href=\"../notes/%v\">note</a> for details ...",
			numDetections, nUUID.String())
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type":                "application/json",
			},
			Body: string(noteLink),
		}, nil
	}

	// In case we haven't detected anything, just confirm receipt and that
	// we were not able to extract any text and hence not taking the note in:
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
			"Content-Type":                "application/json",
		},
		Body: "In the snap provided, we were not able to detect text and hence didn't create a note.",
	}, nil
}

func main() {
	lambda.Start(handler)
}
