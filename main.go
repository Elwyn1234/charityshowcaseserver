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

  http.HandleFunc("/charity-projects", handleCharityProjectsRequest)
  http.HandleFunc("/technologies", handleTechnologiesRequest)
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

func handleTechnologiesRequest(w http.ResponseWriter, r *http.Request) {
  w.Header().Add("Access-Control-Allow-Origin", "*")
  w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
  w.Header().Add("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE")

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
  w.Header().Add("Access-Control-Allow-Origin", "*")
  w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
  w.Header().Add("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE")

  requestBody, _ := ioutil.ReadAll(r.Body)
  switch r.Method {
    case http.MethodPost: {
      createCharityProject(w, requestBody)
    }
    case http.MethodGet: {
      getCharityProjects(w, requestBody)
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

// var page = r.URL.Query()["page"][0]
// var sort = r.URL.Query()["sort"][0]

// TODO: Add a count for how many we want to display on one page
func getCharityProjects(w http.ResponseWriter, requestBody []byte) {
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
      Technologies: make([]Technology, 0),
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

  sqlString := fmt.Sprintf(`UPDATE charityProject SET name='%v', shortDescription='%v', longDescription='%v' WHERE name='%v'`, charityProject.Name, charityProject.ShortDescription, charityProject.LongDescription, charityProject.OldName)
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

type Technology struct {
  Name string
  SVG string
}
type CharityProject struct {
  Name string
  ShortDescription string
  LongDescription string
  Technologies []Technology
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
}

type User struct {
  Username string
  Password string
  Role string
}

// TODO:
// charity-projects/:name/technologies
// Abstract update, create, and delete technologies into one function
// Maybe we should change getCharityProjects function to not return technologies as well
// Use transactions for sql queries
// Maybe open database in scripts folder and call it in install.sh or whatever my script will be called
// Get SQL passwords for root and "ejoh" from a file or other more secure location
