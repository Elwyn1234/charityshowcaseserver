package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"
)

var pool *sql.DB
var logError *log.Logger

type ErrorWriter struct {}
func (errorWriter ErrorWriter) Write(p []byte) (n int, err error) {
  color.Red(string(p))
  return 0, nil
}

func main() {
  var errorWriter ErrorWriter
  logError = log.New(errorWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
  var err error
  pool, err = sql.Open("mysql", "ejoh:YEVT4w2^N4uv2q48TnA9#k&ep@(localhost:3307)/") // TODO: get the password from a file
  if (err != nil) { logError.Panic(err) }
  if err := pool.Ping(); err != nil { logError.Panic(err) }
  _, err = pool.Exec("USE charityShowcase")
  if (err != nil) { logError.Panic(err) }

  http.HandleFunc("/createCharityProject", createCharityProject)
  http.HandleFunc("/getCharityProjects", getCharityProjects)
  http.HandleFunc("/createTechnology", createTechnology)
  http.HandleFunc("/getTechnologies", getTechnologies)
  http.HandleFunc("/login", login)
  http.HandleFunc("/register", register)
  
  log.Print("Listening...")
  log.Print(http.ListenAndServe(":8743", nil))
}

// TODO: Do we need to check the request method before returning? Probably not because we are the only people calling the api and will only call it with one method. log.Print(r.Method)

func register(w http.ResponseWriter, r *http.Request) {
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
  w.Header().Add("Access-Control-Allow-Origin", "*")
  w.Header().Add("Access-Control-Allow-Headers", "Content-Type")

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
  userQueryResult, err := pool.Query(sqlString)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  for (userQueryResult.Next()) {
    if err := userQueryResult.Scan(&user.Username, &user.Password, &user.Role); err != nil {
      logError.Print(err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
  }

  log.Print("login", user)

  var jsonUser, _ = json.Marshal(user)
  w.Write(jsonUser)
}

func createTechnology(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Access-Control-Allow-Origin", "*")
  w.Header().Add("Access-Control-Allow-Headers", "Content-Type")

  requestBody, _ := ioutil.ReadAll(r.Body)
  var technology TechStack
  var err = json.Unmarshal(requestBody, &technology)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  sqlString := fmt.Sprintf(`INSERT INTO techStack (name, imageFileName) VALUES ('%v', '%v')`, technology.Name, technology.SVG)
  log.Print("createTechnology: ", sqlString)
  _, err = pool.Exec(sqlString)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
}

func getTechnologies(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Access-Control-Allow-Origin", "*")

  result, err := pool.Query(`SELECT name FROM techStack`)
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

func createCharityProject(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Access-Control-Allow-Origin", "*")
  w.Header().Add("Access-Control-Allow-Headers", "Content-Type")

  requestBody, _ := ioutil.ReadAll(r.Body)
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

  for i := 0; i < len(charityProject.TechStack); i++ {
    sqlString = fmt.Sprintf(`INSERT INTO techStackToCharityProject (techStack, charityProject) VALUES ('%v', '%v')`, charityProject.TechStack[i].Name, charityProject.Name)
    log.Print("createCharityProject: ", sqlString)
    _, err = pool.Exec(sqlString)
    if (err != nil) {
      logError.Print(err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
  }
}

// TODO: Add a count for how many we want to display on one page
func getCharityProjects(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Access-Control-Allow-Origin", "*")

  // var page = r.URL.Query()["page"][0]
  // var sort = r.URL.Query()["sort"][0]

  charityProjectResult, err := pool.Query(`SELECT * FROM charityProject`)
  if (err != nil) {
    logError.Print(err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  var charityProjects []CharityProject = make([]CharityProject, 0)
  for charityProjectResult.Next() {
    var name string
    var shortDescription string
    var longDescription string
    if err := charityProjectResult.Scan(&name, &shortDescription, &longDescription); err != nil {
      logError.Print(err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
    charityProject := CharityProject {
      Name: name,
      ShortDescription: shortDescription,
      LongDescription: longDescription,
      TechStack: make([]TechStack, 0),
    }

    charityProjects = append(charityProjects, charityProject)
  }

  for i := 0; i < len(charityProjects); i++ {
    sqlString := fmt.Sprintf(`SELECT techStack FROM techStackToCharityProject WHERE charityProject='%v'`, charityProjects[i].Name)
    techStackToCharityProjectResult, err := pool.Query(sqlString)
    if (err != nil) {
      logError.Print(err)
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
    for techStackToCharityProjectResult.Next() {
      var techStackName string
      if err := techStackToCharityProjectResult.Scan(&techStackName); err != nil {
        logError.Print(err)
        w.WriteHeader(http.StatusInternalServerError)
        return
      }

      charityProjects[i].TechStack = append(charityProjects[i].TechStack, TechStack {
        Name: techStackName,
      })
    }
  }
  for charityIndex := 0; charityIndex < len(charityProjects); charityIndex++ {
    for technologyIndex := 0; technologyIndex < len(charityProjects[charityIndex].TechStack); technologyIndex++ {
      technology := charityProjects[charityIndex].TechStack[technologyIndex]
      sqlString := fmt.Sprintf(`SELECT imageFileName FROM techStack WHERE name='%v'`, technology.Name)
      techStackResult := pool.QueryRow(sqlString)
      var techStackImageFileName string
      err = techStackResult.Scan(&techStackImageFileName)
      if (err != nil) {
        logError.Print(err)
        w.WriteHeader(http.StatusInternalServerError)
        return
      }
      technology.SVG = techStackImageFileName
    }
  }

  log.Print("getCharityProjects: ", charityProjects)

  var jsonCharityProjects, _ = json.Marshal(charityProjects)
  w.Write(jsonCharityProjects)
}

type TechStack struct {
  Name string
  SVG string
}
type CharityProject struct {
  Name string
  ShortDescription string
  LongDescription string
  TechStack []TechStack
}

type User struct {
  Username string
  Password string
  Role string
}

// TODO:
// Maybe open database in scripts folder and call it in install.sh or whatever my script will be called
// Get SQL passwords for root and "ejoh" from a file or other more secure location
