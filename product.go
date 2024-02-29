package main

import (
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func productHandler(c *gin.Context) {
	// Example: Extracting data from the request
	name := c.PostForm("name")
	email := c.PostForm("email")
	description := c.PostForm("description")

	// Handle the uploaded file (if any)
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Perform the necessary operations with the file and metadata
	fileName := fmt.Sprintf("%d_%s", header.Size, header.Filename)

	// Generate a unique identifier (UUID) for the product
	productID := uuid.New().String()

	// Upload to S3 and get the generated object key
	objectKey, err := uploadToS3(file, productID, fileName)
	if err != nil {
		fmt.Println("Error uploading file to S3:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading file to S3"})
		return
	}

	// Store metadata in DynamoDB with the unique identifier and S3 object key
	err = storeMetadataInDynamoDB(productID, name, email, description, objectKey)
	if err != nil {
		fmt.Println("Error storing metadata in DynamoDB:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error storing metadata in DynamoDB"})
		return
	}

	// Respond with a success message
	c.JSON(http.StatusOK, gin.H{"message": "Product registered successfully"})
}

func uploadToS3(file multipart.File, productID, originalFileName string) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(AwsRegion),
	})
	if err != nil {
		return "", err
	}

	// Generate a unique identifier (UUID) for the S3 object key
	objectKey := productID + "_" + originalFileName

	s3Client := s3.New(sess)
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(S3ProductBucket),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	// Return the generated object key (UUID + original filename)
	return objectKey, nil
}

func storeMetadataInDynamoDB(productID, name, email, description, objectKey string) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(AwsRegion),
	})
	if err != nil {
		return err
	}

	dynamoDBClient := dynamodb.New(sess)

	_, err = dynamoDBClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(ProductDynamoTable),
		Item: map[string]*dynamodb.AttributeValue{
			"product_id": {
				S: aws.String(productID),
			},
			"name": {
				S: aws.String(name),
			},
			"email": {
				S: aws.String(email),
			},
			"description": {
				S: aws.String(description),
			},
			"image_link": {
				S: aws.String(fmt.Sprintf("https://%s.s3-%s.amazonaws.com/%s", S3ProductBucket, AwsRegion, objectKey)),
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
