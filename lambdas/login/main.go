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
	"github.com/golang-jwt/jwt"
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

      multiValueHeaders, err := login(user)
      if err == sql.ErrNoRows {
        log.Print("Bad credentials")
        response.StatusCode = 400
        response.Body = "{\"success\": false}"
        return response, nil
      }
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 500
        return response, nil
      }

      response.StatusCode = 200
      response.Body = "{\"success\": true}"
      response.MultiValueHeaders = multiValueHeaders
    }



    default: {
      log.Print("Invalid HTTP Method")
      response.StatusCode = 400
      return response, nil
    }
  }



  response.StatusCode = 200
  log.Print("handler END ")
  return response, nil
}






func login(user User) (map[string][]string, error) {
  sqlString := fmt.Sprintf(`SELECT * FROM user WHERE username='%v' AND password='%v'`, user.Username, user.Password)
  userQueryResult := pool.QueryRow(sqlString)
  err := userQueryResult.Scan(&user.Username, &user.Password, &user.Role)
  if err != nil {
    return nil, err
  }

  log.Print("login", user)

  multiValueHeaders, err := setCookies(user)
  if err != nil { return nil, err } // TODO: maybe set success: false

  return multiValueHeaders, nil
}






func setCookies(user User) (map[string][]string, error) {
  token := jwt.New(jwt.SigningMethodRS256)
  claims := token.Claims.(jwt.MapClaims)
  // claims["exp"] = time.Now().Add(10 * time.Minute) // TODO: why does this make the token invalid
  // TODO: have a short expiry time and rotate them frequently (note: there is no way to invalidate a token)
  claims["sub"] = user.Username
  claims["role"] = user.Role // TODO: set a "role" cookie as well and update the client to display content based on this. Similar to the loggedIn cookie
  tokenString, err := token.SignedString(privateKey)
  if err != nil {
    return nil, err
  }

  cookies := make([]string, 3)
  setCookieHeaderValue := fmt.Sprintf(`jwt=%v; Path=/; Domain=localhost; SameSite=None; Secure`, tokenString)
  cookies[0] = setCookieHeaderValue // TODO: will the jwt ever contain illegal characters 
  cookies[1] = "loggedIn=true" // We do this to give the client an easy way to check if the user is logged in. This is ideal because the lifetime and scope of this cookie will match that of the jwt cookie, giving an accurate representation of whether the user is logged in. This is unlike browser memory, sessionStorage, and localStorage which all have different variable lifetimes and scope. Check this variable because TODO: jwt cookie will have httpOnly set, making it unreachable by js code.
  cookies[2] = fmt.Sprintf("role=%v", user.Role)

  multiValueHeaders := make(map[string][]string)
  multiValueHeaders["Set-Cookie"] = cookies

  for outer_key, outer_value := range multiValueHeaders {
    log.Print("outer_key: ", outer_key)
    for _, inner_value := range outer_value {
      log.Print("inner_value: ", inner_value)
    }
  }

  return multiValueHeaders, nil
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
