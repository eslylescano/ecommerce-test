package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// User model
type User struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// JWT claims struct
type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

func registerHandler(c *gin.Context) {
	var user User

	// Bind JSON request to User struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the password
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		fmt.Print(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Save user data to DynamoDB
	if err := saveUserToDynamoDB(user.Email, user.Name, hashedPassword); err != nil {
		fmt.Print(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user data"})
		return
	}

	// Create JWT token
	token, err := createToken(user.Email)
	if err != nil {
		fmt.Print(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JWT token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	fmt.Print(err)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func saveUserToDynamoDB(email, name, password string) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(AwsRegion),
	})
	if err != nil {
		fmt.Print(err)
		return err
	}

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Marshal the User struct into DynamoDB attribute values
	userItem, err := dynamodbattribute.MarshalMap(User{
		Email:    email,
		Name:     name,
		Password: password,
	})
	if err != nil {
		fmt.Print(err)
		return err
	}

	// Put item into DynamoDB
	input := &dynamodb.PutItemInput{
		Item:      userItem,
		TableName: aws.String(UserDynamoTable),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		fmt.Print(err)
		return err
	}

	return nil
}

func createToken(email string) (string, error) {
	claims := Claims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(SecretKey)
	if err != nil {
		fmt.Print(err)
		return "", err
	}

	return signedToken, nil
}
