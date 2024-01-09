package main

import (
  "log"
  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
)

func main() {
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



  switch request.HTTPMethod {
    case "POST": {
      multiValueHeaders, err := logout()
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 500
        return response, nil
      }

      response.MultiValueHeaders = multiValueHeaders
    }



    default: {
      log.Print("Invalid HTTP Method")
      response.StatusCode = 400
      return response, nil
    }
  }



  return response, nil
}






func logout() (map[string][]string, error) {
  // Logging out simply sets the client jwt and loggedIn cookies to an empty string
  cookies := make([]string, 3)
  cookies[0] = "jwt="
  cookies[1] = "loggedIn=false"
  cookies[2] = "role="

  multiValueHeaders := make(map[string][]string)
  multiValueHeaders["Set-Cookie"] = cookies

  log.Print("logout succesful")

  return multiValueHeaders, nil
}

