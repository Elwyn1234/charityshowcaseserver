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
      users, err := getUsers()
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 500
        return response, nil
      }

      responseBodyBytes, err := json.Marshal(users)
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

      var user User
      err = json.Unmarshal([]byte(decodedBody), &user)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 400
        return response, nil
      }

      err = postUser(user)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 500
        return response, nil
      }
    }



    case "PUT": {
      decodedBody, err := base64.StdEncoding.DecodeString(request.Body)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 400
        return response, nil
      }

      var user User
      err = json.Unmarshal([]byte(decodedBody), &user)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 400
        return response, nil
      }

      err = putUser(user)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 500
        return response, nil
      }
    }



    case "DELETE": {
      decodedBody, err := base64.StdEncoding.DecodeString(request.Body)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 400
        return response, nil
      }

      var user User
      err = json.Unmarshal([]byte(decodedBody), &user)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 400
        return response, nil
      }

      err = deleteUser(user.Username)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 500
        return response, nil
      }
    }



    default: {
      log.Print("Invalid HTTP Method")
      response.StatusCode = 400
      return response, nil
    }
  }



  return response, nil
}






func getUsers() ([]User, error) {
  var users []User = make([]User, 0)

  result, err := pool.Query(`SELECT username, role FROM user`)
  if (err != nil) {
    // logError.Print(err)
    return users, err
  }

  for result.Next() {
    var user User;
    if err := result.Scan(&user.Username, &user.Role); err != nil {
      return users, err
    }
    users = append(users, user)
  }
  log.Print("getUsers: ", users)

  return users, nil
}






func postUser(user User) (error) {
  sqlString := fmt.Sprintf(`INSERT INTO user
      (username, password, role)
      VALUES ('%v', '%v', '%v')`,
      user.Username,
      user.Password,
      user.Role)

      _, err := pool.Exec(sqlString)
  if (err != nil) {
    return err
  }

  return nil
}






func putUser(user User) (error) {
 // TODO: currently only supports updating the role, lets add more functionality when needed by the front end
  _, err := pool.Exec(`UPDATE user SET role=? WHERE username=?`,
    user.Role,
    user.Username)

  if (err != nil) {
    return err
  }

  return nil
}






func deleteUser(username string) (error) {
  _, err := pool.Exec(`DELETE from user WHERE username=?`,
    username)

  if (err != nil) {
    return err
  }

  return nil
}





var pool *sql.DB
var privateKey *rsa.PrivateKey

type User struct {
  Username string
  Password string
  Role string
}






// func validateJwt(w http.ResponseWriter, r *http.Request) (isValid bool) {
//   // Auth
//   jwtString, err := r.Cookie("jwt")
//   if err != nil {
//     logError.Print(err)
//     w.WriteHeader(http.StatusUnauthorized)
//     return
//   }

//   // token, err := regexp.Compile("^Bearer ") Do we need to prepend Bearer to the token when sending it to the server?!?!
//   // TODO: use 'ok' instead of 'err' for variable name
//   token, err := jwt.Parse(jwtString.Value, func(token *jwt.Token) (interface{}, error) {
//     return &privateKey.PublicKey, nil
//   })
//   if err != nil {
//     logError.Print(err)
//     w.WriteHeader(http.StatusUnauthorized)
//     return
//   }
//   _, ok := token.Method.(*jwt.SigningMethodRSA) // TODO: Why do we need to check the method type
//   if !ok {
//     logError.Print(err)
//     w.WriteHeader(http.StatusUnauthorized)
//     return false // TODO: Write an informative error messages?!
//   }
//   if !token.Valid {
//     logError.Print(err)
//     w.WriteHeader(http.StatusUnauthorized)
//     return false
//   }
//   return true
// }
