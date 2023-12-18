package main

import (
	"database/sql"
  "log"
	"encoding/json"
	"encoding/base64"
  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
  var err error
  pool, err = sql.Open("mysql", "admin:ealkjwahebf@tcp(charity-showcase-database-mysql.cb6tbxpuewpo.eu-north-1.rds.amazonaws.com:3306)/charityshowcase") // TODO: get the password from a file
  if (err != nil) { log.Panic(err) }

  lambda.Start(handler)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
  response := events.APIGatewayProxyResponse {
    StatusCode: 200,
  }


  response.Headers = make(map[string]string)
  response.Headers["Access-Control-Allow-Origin"] = "*"
  response.Headers["Access-Control-Allow-Headers"] = "Content-Type"
  response.Headers["Access-Control-Allow-Credentials"] = "true"

  decodedBody, err := base64.StdEncoding.DecodeString(request.Body)
  if err != nil {
    log.Print(err)
    response.StatusCode = 400
    return response, nil
  }
  log.Print("decodedBody")
  log.Print(decodedBody)

  var user User
  err = json.Unmarshal([]byte(decodedBody), &user)
  if (err != nil) {
    log.Print(err)
    response.StatusCode = 400
    return response, nil
  }

  log.Print("Username: ", user.Username)
  log.Print("Password: ", user.Password)
  log.Print("Role: ", user.Role)

  responseBodyBytes, err := json.Marshal(user)
  if (err != nil) {
    log.Print(err)
    response.StatusCode = 500
    return response, nil
  }
  response.Body = string(responseBodyBytes[:])

  return response, nil
}

var pool *sql.DB

type User struct {
  Username string
  Password string
  Role string
}
