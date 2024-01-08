package main

import (
  "fmt"
	"crypto/rand"
	"crypto/rsa"
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
  pool, err = sql.Open("mysql", "admin:ealkjwahebf@tcp(charity-showcase.cos7ursa6kc8.eu-west-2.rds.amazonaws.com:3306)/charityshowcase") // TODO: get the password from a file
  if (err != nil) { log.Panic(err) }
  err = pool.Ping()
  if (err != nil) { log.Panic(err) }

  privateKey, err = rsa.GenerateKey(rand.Reader, 2048)

  lambda.Start(handler)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
  response := events.APIGatewayProxyResponse {
    StatusCode: 200,
  }


  // TODO:
  // w.Header().Add("Access-Control-Allow-Origin", "http://localhost:3000")
  response.Headers = make(map[string]string)
  response.Headers["Access-Control-Allow-Origin"] = "*"
  response.Headers["Access-Control-Allow-Headers"] = "Content-Type"
  response.Headers["Access-Control-Allow-Credentials"] = "true"
  response.Headers["Access-Control-Allow-Methods"] = "POST, GET, PUT, DELETE"
  if (request.HTTPMethod == "OPTIONS") {
    return response, nil
  }

  // if !validateJwt(w, r) { return }



  switch request.HTTPMethod {
    case "GET": {
      technologies, err := getTechnologies()
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 500
        return response, nil
      }

      responseBodyBytes, err := json.Marshal(technologies)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 500
        return response, nil
      }

      response.Body = string(responseBodyBytes[:])
    }



    case "POST": {
      decodedBody, err := base64.StdEncoding.DecodeString(request.Body)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 400
        return response, nil
      }

      var technology Technology
      err = json.Unmarshal([]byte(decodedBody), &technology)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 400
        return response, nil
      }

      err = postTechnology(technology)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 500
        return response, nil
      }
    }



    // case "PUT": {
    // }



    // case "DELETE": {
    // }



    default: {
      log.Print("Invalid HTTP Method")
      response.StatusCode = 400
      return response, nil
    }
  }



  return response, nil
}






func getTechnologies() ([]string, error) {
  var technologies []string = make([]string, 0)

  result, err := pool.Query(`SELECT name FROM technology`)
  if (err != nil) {
    return technologies, err
  }

  for result.Next() {
    var name string
    if err := result.Scan(&name); err != nil {
      return technologies, err
    }
    technologies = append(technologies, name)
  }
  log.Print("getTechnologies: ", technologies)

  return technologies, nil
}







func postTechnology(technology Technology) (error) {
  sqlString := fmt.Sprintf(`INSERT INTO technology
      (name, imageFileName)
      VALUES ('%v', '%v')`,
      technology.Name,
      technology.SVG)

      _, err := pool.Exec(sqlString)
  if (err != nil) {
    return err
  }

  return nil
}






var pool *sql.DB
var privateKey *rsa.PrivateKey

type Technology struct {
  Name string
  SVG string
}
