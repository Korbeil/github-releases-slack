package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"bytes"
)

func handle(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse needed values from GitHub webhook payload
	request, err := parseRequest(req)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Unable to handle request", StatusCode: 500}, nil
	}

	// Create message from request
	message, err := messageFromRequest(*request)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Unable to create message", StatusCode: 500}, nil
	}

	// Send request to Slack API
	response, err := postMessageToSlack(message, request.Token)
	if err != nil || !response.OK {
		buffer := bytes.NewBufferString("Unable to send request to Slack - ")
		buffer.Write(message)
		return events.APIGatewayProxyResponse{Body: buffer.String(), StatusCode: 500}, nil
	}

	// Send response
	return events.APIGatewayProxyResponse{Body: "{ \"done\": true }", StatusCode: 200}, nil
}

func main() {
	lambda.Start(handle)
}
