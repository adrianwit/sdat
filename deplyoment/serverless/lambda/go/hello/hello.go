package main

import "github.com/aws/aws-lambda-go/lambda"

type Person struct {
	Name string
}

func handle() (string, error) {
	return "Hello World", nil
}

func main() {
	lambda.Start(handle)
}
