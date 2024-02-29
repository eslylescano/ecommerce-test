package main

import "os"

var AwsRegion = getEnv("AWS_REGION", "eu-west-2")
var S3ProductBucket = getEnv("S3_BUCKET", "ecomerce-test-esly")
var ProductDynamoTable = getEnv("PRODUCT_DYNAMODB_TABLE", "products")
var UserDynamoTable = getEnv("USER_DYNAMODB_TABLE", "users-ecomerce-test")
var SecretKey = []byte("your-secret-key")

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
