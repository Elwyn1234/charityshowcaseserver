package main

import (
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt"
)

func main() {
  var errorWriter ErrorWriter
  logError = log.New(errorWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
  var err error
  pool, err = sql.Open("mysql", "ejoh:YEVT4w2^N4uv2q48TnA9#k&ep@(localhost:3307)/charityShowcase") // TODO: get the password from a file
  if (err != nil) { logError.Panic(err) }
  err = pool.Ping()
  if (err != nil) { logError.Panic(err) }

  privateKey, err = rsa.GenerateKey(rand.Reader, 2048)

  http.HandleFunc("/charity-projects/", handleCharityProjectsRequest)
  http.HandleFunc("/technologies", handleTechnologiesRequest)
  http.HandleFunc("/login", login)
  http.HandleFunc("/logout", logout)
  http.HandleFunc("/register", register)
  
  log.Print("Listening...")
  log.Print(http.ListenAndServe(":8743", nil))
}

func register(w http.ResponseWriter, r *http.Request) {
  // TODO: Do we need to check the request method before returning? Probably not because we are the only people calling the api and will only call it with one method. log.Print(r.Method)

  w.Header().Add("Access-Control-Allow-Origin", "*")
  w.Header().Add("Access-Control-Allow-Headers", "Content-Type")

  requestBody, _ := ioutil.ReadAll(r.Body)
  var user User
  var err = json.Unmarshal(requestBody, &user)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  sqlString := fmt.Sprintf(`INSERT INTO user (username, password, role) VALUES ('%v', '%v', '%v')`, user.Username, user.Password, user.Role)
  log.Print("register: ", sqlString)
  _, err = pool.Exec(sqlString)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
}

func login(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Access-Control-Allow-Origin", "http://localhost:3000")
  w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
  w.Header().Add("Access-Control-Allow-Credentials", "true")

  requestBody, _ := ioutil.ReadAll(r.Body)
  log.Print(requestBody)
  var user User
  var err = json.Unmarshal(requestBody, &user)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  sqlString := fmt.Sprintf(`SELECT * FROM user WHERE username='%v' AND password='%v'`, user.Username, user.Password)
  log.Print("login: ", sqlString)
  userQueryResult := pool.QueryRow(sqlString)
  err = userQueryResult.Scan(&user.Username, &user.Password, &user.Role)
  if err == sql.ErrNoRows {
    logError.Print("Bad credentials")
    w.Write([]byte("{\"success\": false}"))
    return
  }
  if err != nil {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  log.Print("login", user)
  if setCookies(w, user) != nil { return } // TODO: maybe set success: false
  w.Write([]byte("{\"success\": true}"))
}

func logout(w http.ResponseWriter, r *http.Request) {
  // Logging out simply sets the client jwt and loggedIn cookies to an empty string
  w.Header().Add("Access-Control-Allow-Origin", "http://localhost:3000")
  w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
  w.Header().Add("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE")
  w.Header().Add("Access-Control-Allow-Credentials", "true")

  w.Header().Add("Set-Cookie", "jwt=")
  w.Header().Add("Set-Cookie", "loggedIn=false")
  log.Print("logout succesful")
}

func setCookies(w http.ResponseWriter, user User) (error) {
  token := jwt.New(jwt.SigningMethodRS256)
  claims := token.Claims.(jwt.MapClaims)
  // claims["exp"] = time.Now().Add(10 * time.Minute) // TODO: why does this make the token invalid
  // TODO: have a short expiry time and rotate them frequently (note: there is no way to invalidate a token)
  claims["sub"] = user.Username
  claims["role"] = user.Role // TODO: set a "role" cookie as well and update the client to display content based on this. Similar to the loggedIn cookie
  tokenString, err := token.SignedString(privateKey)
  if err != nil {
    logError.Print(err)
    w.WriteHeader(http.StatusUnauthorized) // TODO: is this the correct status
    return err
  }
  setCookieHeaderValue := fmt.Sprintf(`jwt=%v; Path=/; Domain=localhost; SameSite=None; Secure`, tokenString)
  w.Header().Add("Set-Cookie", setCookieHeaderValue) // TODO: will the jwt ever contain illegal characters 
  w.Header().Add("Set-Cookie", "loggedIn=true") // We do this to give the client an easy way to check if the user is logged in. This is ideal because the lifetime and scope of this cookie will match that of the jwt cookie, giving an accurate representation of whether the user is logged in. This is unlike browser memory, sessionStorage, and localStorage which all have different variable lifetimes and scope. Check this variable because TODO: jwt cookie will have httpOnly set, making it unreachable by js code.
  return nil
}

func validateJwt(w http.ResponseWriter, r *http.Request) (isValid bool) {
  // Auth
  jwtString, err := r.Cookie("jwt")
  if err != nil {
    logError.Print(err)
    w.WriteHeader(http.StatusUnauthorized)
    return
  }

  // token, err := regexp.Compile("^Bearer ") Do we need to prepend Bearer to the token when sending it to the server?!?!
  // TODO: use 'ok' instead of 'err' for variable name
  token, err := jwt.Parse(jwtString.Value, func(token *jwt.Token) (interface{}, error) {
    return &privateKey.PublicKey, nil
  })
  if err != nil {
    logError.Print(err)
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
  _, ok := token.Method.(*jwt.SigningMethodRSA) // TODO: Why do we need to check the method type
  if !ok {
    logError.Print(err)
    w.WriteHeader(http.StatusUnauthorized)
    return false // TODO: Write an informative error messages?!
  }
  if !token.Valid {
    logError.Print(err)
    w.WriteHeader(http.StatusUnauthorized)
    return false
  }
  return true
}

func handleTechnologiesRequest(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Access-Control-Allow-Origin", "http://localhost:3000")
  w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
  w.Header().Add("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE")
  w.Header().Add("Access-Control-Allow-Credentials", "true")

  if !validateJwt(w, r) { return }

  requestBody, _ := ioutil.ReadAll(r.Body)
  switch r.Method {
    case http.MethodPost: {
      createTechnology(w, requestBody)
    }
    case http.MethodGet: {
      getTechnologies(w, requestBody)
    }
    case http.MethodPut: {
    }
    case http.MethodDelete: {
    }
  }
}

func createTechnology(w http.ResponseWriter, requestBody []byte) {
  var technology Technology
  var err = json.Unmarshal(requestBody, &technology)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  sqlString := fmt.Sprintf(`INSERT INTO technology (name, imageFileName) VALUES ('%v', '%v')`, technology.Name, technology.SVG)
  log.Print("createTechnology: ", sqlString)
  _, err = pool.Exec(sqlString)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
}

func getTechnologies(w http.ResponseWriter, requestBody []byte) {
  result, err := pool.Query(`SELECT name FROM technology`)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  var technologies []string = make([]string, 0)
  for result.Next() {
    var name string
    if err := result.Scan(&name); err != nil {
      logError.Print(err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
    technologies = append(technologies, name)
  }
  log.Print("getTechnologies: ", technologies)

  var jsonTechnologies, _ = json.Marshal(technologies)
  w.Write(jsonTechnologies)
}

func handleCharityProjectsRequest(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Access-Control-Allow-Origin", "http://localhost:3000")
  w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
  w.Header().Add("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE")
  w.Header().Add("Access-Control-Allow-Credentials", "true")

  if !validateJwt(w, r) { 
    log.Print("not a vaild token")
    return }

  requestBody, _ := ioutil.ReadAll(r.Body)
  switch r.Method {
    case http.MethodPost: {
      createCharityProject(w, requestBody)
    }
    case http.MethodGet: {
      pathSegments := strings.Split(r.URL.Path, "/")
      if (len(pathSegments) <= 2 || pathSegments[2] == "") {
        getCharityProjects(w)
      } else {
        getCharityProject(w, pathSegments[2]) // matches /charity-projects/:name
      }
    }
    case http.MethodPut: {
      updateCharityProject(w, requestBody)
    }
    case http.MethodDelete: {
    }
  }
}

func createCharityProject(w http.ResponseWriter, requestBody []byte) {
  var charityProject CharityProject
  var err = json.Unmarshal(requestBody, &charityProject)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  sqlString := fmt.Sprintf(`INSERT INTO charityProject (name, shortDescription, longDescription) VALUES ('%v', '%v', '%v')`, charityProject.Name, charityProject.ShortDescription, charityProject.LongDescription)
  log.Print("createCharityProject: ", sqlString)
  _, err = pool.Exec(sqlString)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  for i := 0; i < len(charityProject.Technologies); i++ {
    sqlString = fmt.Sprintf(`INSERT INTO technologyToCharityProject (technology, charityProject) VALUES ('%v', '%v')`, charityProject.Technologies[i].Name, charityProject.Name)
    log.Print("createCharityProject: ", sqlString)
    _, err = pool.Exec(sqlString)
    if (err != nil) {
      logError.Print(err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
  }
}

func getCharityProject(w http.ResponseWriter, charityProjectName string) {
// var page = r.URL.Query()["page"][0]
// var sort = r.URL.Query()["sort"][0]
  sqlString := fmt.Sprintf(`SELECT * FROM charityProject WHERE name='%v'`, charityProjectName)
  charityProjectResult := pool.QueryRow(sqlString)
  charityProject := CharityProject {
    Technologies: make([]Technology, 0),
  }
  err := charityProjectResult.Scan(&charityProject.Name, &charityProject.ShortDescription, &charityProject.LongDescription, &charityProject.Archived)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  sqlString = fmt.Sprintf(`SELECT technology FROM technologyToCharityProject WHERE charityProject='%v'`, charityProject.Name)
  technologyToCharityProjectResult, err := pool.Query(sqlString)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  for technologyToCharityProjectResult.Next() {
    var technologyName string
    if err := technologyToCharityProjectResult.Scan(&technologyName); err != nil {
      logError.Print(err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }

    charityProject.Technologies = append(charityProject.Technologies, Technology {
      Name: technologyName,
    })
  }

  for i := 0; i < len(charityProject.Technologies); i++ {
    technology := charityProject.Technologies[i]
    sqlString := fmt.Sprintf(`SELECT imageFileName FROM technology WHERE name='%v'`, technology.Name)
    technologyResult := pool.QueryRow(sqlString)
    var technologyImageFileName string
    err = technologyResult.Scan(&technologyImageFileName)
    if (err != nil) {
      logError.Print(err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
    technology.SVG = technologyImageFileName
  }

  log.Print("getCharityProject: ", charityProject)

  var jsonCharityProjects, _ = json.Marshal(charityProject)
  w.Write(jsonCharityProjects)
}

func getCharityProjects(w http.ResponseWriter) {
// TODO: Add a count for how many we want to display on one page
  charityProjectResult, err := pool.Query(`SELECT * FROM charityProject`)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  var charityProjects []CharityProject = make([]CharityProject, 0)
  for charityProjectResult.Next() {
    charityProject := CharityProject {
      Technologies: make([]Technology, 0),
    }
    err = charityProjectResult.Scan(&charityProject.Name, &charityProject.ShortDescription, &charityProject.LongDescription, &charityProject.Archived)
    if err != nil {
      logError.Print(err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }

    charityProjects = append(charityProjects, charityProject)
  }

  for i := 0; i < len(charityProjects); i++ {
    sqlString := fmt.Sprintf(`SELECT technology FROM technologyToCharityProject WHERE charityProject='%v'`, charityProjects[i].Name)
    technologyToCharityProjectResult, err := pool.Query(sqlString)
    if (err != nil) {
      logError.Print(err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
    for technologyToCharityProjectResult.Next() {
      var technologyName string
      if err := technologyToCharityProjectResult.Scan(&technologyName); err != nil {
        logError.Print(err)
        w.WriteHeader(http.StatusInternalServerError)
        return
      }

      charityProjects[i].Technologies = append(charityProjects[i].Technologies, Technology {
        Name: technologyName,
      })
    }
  }
  for charityIndex := 0; charityIndex < len(charityProjects); charityIndex++ {
    for technologyIndex := 0; technologyIndex < len(charityProjects[charityIndex].Technologies); technologyIndex++ {
      technology := charityProjects[charityIndex].Technologies[technologyIndex]
      sqlString := fmt.Sprintf(`SELECT imageFileName FROM technology WHERE name='%v'`, technology.Name)
      technologyResult := pool.QueryRow(sqlString)
      var technologyImageFileName string
      err = technologyResult.Scan(&technologyImageFileName)
      if (err != nil) {
        logError.Print(err)
        w.WriteHeader(http.StatusInternalServerError)
        return
      }
      technology.SVG = technologyImageFileName
    }
  }

  log.Print("getCharityProjects: ", charityProjects)

  var jsonCharityProjects, _ = json.Marshal(charityProjects)
  w.Write(jsonCharityProjects)
}

func updateCharityProject(w http.ResponseWriter, requestBody []byte) {
  var charityProject CharityProjectUpdate
  var err = json.Unmarshal(requestBody, &charityProject)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  sqlString := fmt.Sprintf(`UPDATE charityProject SET name=COALESCE('%v', name), shortDescription=COALESCE('%v', shortDescription), longDescription=COALESCE('%v', longDescription), archived=COALESCE(%v, archived) WHERE name='%v'`, charityProject.Name, charityProject.ShortDescription, charityProject.LongDescription, charityProject.Archived, charityProject.OldName)
  log.Print("updateCharityProject: ", sqlString)
  _, err = pool.Exec(sqlString)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  for i := 0; i < len(charityProject.Technologies); i++ {
    sqlString = fmt.Sprintf(`UPDATE technologyToCharityProject SET technology='%v' WHERE technology='%v' and charityProject='%v'`, charityProject.Technologies[i].Name, charityProject.Technologies[i].OldName, charityProject.Name)
    log.Print("updateCharityProject: ", sqlString)
    _, err = pool.Exec(sqlString)
    if (err != nil) {
      logError.Print(err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
  }
}

func (errorWriter ErrorWriter) Write(p []byte) (n int, err error) {
  color.Red(string(p))
  return 0, nil
}

type Technology struct {
  Name string
  SVG string
}
type CharityProject struct {
  Name string
  ShortDescription string
  LongDescription string
  Technologies []Technology
  Archived bool
}
type TechnologyUpdate struct {
  OldName string
  Name string
  SVG string
}
type CharityProjectUpdate struct {
  OldName string
  Name string
  ShortDescription string
  LongDescription string
  Technologies []TechnologyUpdate
  Archived bool
}
type User struct {
  Username string
  Password string
  Role string
}
type ErrorWriter struct {}

// Globals
var pool *sql.DB
var logError *log.Logger
var privateKey *rsa.PrivateKey

// TODO:
// Don't use the default logger
// charity-projects/:name/technologies
// Abstract update, create, and delete technologies into one function
// Maybe we should change getCharityProjects function to not return technologies as well
// Use transactions for sql queries
// Maybe open database in scripts folder and call it in install.sh or whatever my script will be called
// Get SQL passwords for root and "ejoh" from a file or other more secure location

// SECURITY considerations:
// TODO: CSRF will be mitigated by the fact that Access-Control-Allow-Origin will be set to the react app's domain. This means that only the react app can make requests to the Go server. CSRF could be used to hit a page of our React app such as /edit-charity-project/:name however this would only get and display the front end form to the victim (the user) but would not make any POST, PUT, or DELETE requests to the Go server where our database is. Therefore, the victim is protected from CSRF. This assumes that hitting any URL in our React app does not make automatic requests to the Go server. For example, hitting /delete-charity-project/:name should not immdediately delete the Charity Project record but should instead prompt the user to confirm deletion. This means that all actions require user confirmation or action and no CSRF is possible. Does CORS prevent non browser requests such as curl requests from impersonating a domain? At the moment, we are storing the JWT as a cookie and other sites will have access to this cookie but we are relying on the Go server to check that requests are being made by the React app's domain to stop other domains from making requests with the JWT cookie. Because of the above, I do not believe that we need to leverage the double submit cookie pattern.

// DONE:
  // TODO: Generate key pair
  // TODO: make client store the token
  // TODO: make client pass the token in for all requests
