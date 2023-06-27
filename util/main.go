package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"
)

func stringReplaceFirstGroup(regex string, str string, repl string) (newstr string) {
  compiledRegex, _ := regexp.Compile(regex)
  matchIndexes := compiledRegex.FindStringSubmatchIndex(str)
  start := matchIndexes[2]
  end := matchIndexes[3]
  startSlice := str[0:start]
  endSlice := str[end:]
  return startSlice + repl + endSlice
}

func main() {
  var errorWriter ErrorWriter
  logError = log.New(errorWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

  if (len(os.Args) < 2) {
    logError.Fatal("No subcommands specified. Exiting.")
  }
  
  createdbFlags := flag.NewFlagSet("createdb", flag.PanicOnError)
  var createdbFlagsForce bool
  createdbFlags.BoolVar(&createdbFlagsForce, "f", false, "")
  createdbFlags.BoolVar(&createdbFlagsForce, "force", false, "Force the entire database to be dropped and recreated.")

  installFlags := flag.NewFlagSet("install", flag.PanicOnError)
  var installFlagsForce bool
  installFlags.BoolVar(&installFlagsForce, "f", false, "")
  installFlags.BoolVar(&installFlagsForce, "force", false, "Force the entire database to be dropped and recreated.")

  switch os.Args[1] {
  case "install":
    err := installFlags.Parse(os.Args[2:])
    if (err != nil) { log.Print(err.Error()) }

    log.Print("install flags force: ", installFlagsForce)
    createdb(installFlagsForce)
    addTestData()

  case "createdb":
    err := createdbFlags.Parse(os.Args[2:])
    if (err != nil) { log.Print(err.Error()) }
    createdb(createdbFlagsForce)

  case "dropdb":
    dropDatabase(rootOpenDatabase(""))
  default:
    logError.Fatal("The subcommand provided is not a recognised subcommand. Exiting.")
  }
}

func rootOpenDatabase(dbname string) (pool *sql.DB) {
  pool, err := sql.Open("mysql", fmt.Sprintf("admin:ealkjwahebf@tcp(charity-showcase-database-mysql.cb6tbxpuewpo.eu-north-1.rds.amazonaws.com:3306)/%v", dbname)) // TODO: get the password from a file
  if (err != nil) { logError.Fatal(err.Error()) } // TODO: error handling

  pool.SetConnMaxLifetime(0)
  pool.SetMaxIdleConns(1)
  pool.SetMaxOpenConns(1)
  if err := pool.Ping(); err != nil { logError.Fatal(err.Error()) }

  return pool
}

func dropDatabase(pool *sql.DB) {
  log.Print("Dropping charityshowcase Database.")
  _, err := pool.Exec("DROP DATABASE charityshowcase")
  if (err != nil) { logError.Fatal(err.Error()) }
  log.Print("Database dropped!")
}

func createdb(forceCreation bool) {
  pool := rootOpenDatabase("")
  if (forceCreation){
    dropDatabase(pool)
  }
  log.Print("Creating charityshowcase Database.")
  _, err := pool.Exec(`CREATE DATABASE charityshowcase;`)
  if (err != nil) { logError.Fatal(err.Error()) }
  log.Print("charityshowcase database successfully created!")
  _, err = pool.Exec(`USE charityshowcase;`)
  if (err != nil) { logError.Fatal(err.Error()) }
  log.Print("charityshowcase database selected.")
    
  _, err = pool.Exec(`CREATE TABLE charityProject (
    name VARCHAR(50) NOT NULL,
    shortDescription VARCHAR(300) NOT NULL,
    longDescription VARCHAR(5000),
    charityName VARCHAR(200),
    charityEmail VARCHAR(100),
    projectEmail VARCHAR(100) NOT NULL,
    location VARCHAR(200) NOT NULL,
    archived BOOL NOT NULL DEFAULT false,
    PRIMARY KEY (name)
  );`)
  if (err != nil) { logError.Fatal(err.Error()) }
  log.Print("charityProject table created!")
    
  _, err = pool.Exec(`CREATE TABLE technology (
    name VARCHAR(32) NOT NULL,
    imageFileName VARCHAR(64),
    PRIMARY KEY (name)
  )`)
  if (err != nil) { logError.Fatal(err.Error()) }
  log.Print("technology table created!")

  _, err = pool.Exec(`CREATE TABLE technologyToCharityProject (
    technology VARCHAR(50) NOT NULL,
    charityProject VARCHAR(50) NOT NULL,
    PRIMARY KEY (technology, charityProject),
    FOREIGN KEY (technology)
      REFERENCES technology(name)
      ON DELETE CASCADE
      ON UPDATE CASCADE,
    FOREIGN KEY (charityProject)
      REFERENCES charityProject(name)
      ON DELETE CASCADE
      ON UPDATE CASCADE
  );`)
  if (err != nil) { logError.Fatal(err.Error()) }
  log.Print("technologyToCharityProject table created!")

  _, err = pool.Exec(`CREATE TABLE user (
    username VARCHAR(32) NOT NULL,
    password VARCHAR(32) NOT NULL,
    role ENUM('user', 'editor', 'creator', 'admin'),
    PRIMARY KEY (username)
  );`)
  if (err != nil) { logError.Fatal(err.Error()) }
  log.Print("user table created!")
}

func addAdminUser() {
  pool := rootOpenDatabase("charityshowcase")
  _, err := pool.Exec(`INSERT INTO user (username, password, role) VALUES ("admin", "pass", "admin");`) // TODO: more secure credentials
  if (err != nil) { logError.Fatal(err.Error()) }
  log.Print("Admin user created!")
}

func addTestData() {
  pool := rootOpenDatabase("charityshowcase")
  
  testdata, err := os.ReadFile("./testdata.json")
  if (err != nil) {
    logError.Fatal("Failed to read file testdata.json")
  }
  var charityShowcase CharityShowcase
  err = json.Unmarshal(testdata, &charityShowcase)
  if (err != nil) {
    logError.Fatal(err)
  }

  for i := 0; i < len(charityShowcase.Technologies); i++ {
    _, err = pool.Exec(`INSERT INTO technology (name, imageFileName) VALUES (?, ?);`, charityShowcase.Technologies[i].Name, charityShowcase.Technologies[i].SVG) // TODO: more secure credentials
    if (err != nil) { logError.Fatal(err.Error()) }
  }
  log.Print("Test data created for the technology table!")

  for i := 0; i < len(charityShowcase.CharityProjects); i++ {
    charityProject := charityShowcase.CharityProjects[i]
    _, err = pool.Exec(`INSERT INTO charityProject (name, shortDescription, longDescription, charityName, charityEmail, projectEmail, location, archived) VALUES (?, ?, ?, ?, ?, ?, ?, ?);`, charityProject.Name, charityProject.ShortDescription, charityProject.LongDescription, charityProject.CharityName, charityProject.CharityEmail, charityProject.ProjectEmail, charityProject.Location, charityProject.Archived) // TODO: more secure credentials
    if (err != nil) { logError.Fatal(err.Error()) }
    for technologyIndex := 0; technologyIndex < len(charityProject.Technologies); technologyIndex++ {
      technology := charityProject.Technologies[technologyIndex]
      _, err = pool.Exec(`INSERT INTO technologyToCharityProject (technology, charityProject) VALUES (?, ?);`, technology.Name, charityProject.Name) // TODO: more secure credentials
      if (err != nil) { logError.Fatal(err.Error()) }
    }
  }
  log.Print("Test data created for the charityproject table!")

  for i := 0; i < len(charityShowcase.Users); i++ {
    user := charityShowcase.Users[i]
    _, err = pool.Exec(`INSERT INTO user (username, password, role) VALUES (?, ?, ?);`, user.Username, user.Password, user.Role) // TODO: more secure credentials
    if (err != nil) { logError.Fatal(err.Error()) }
  }
  log.Print("Test data created for the user table!")
  addAdminUser()
}

type CharityShowcase struct {
  Technologies []Technology
  CharityProjects []CharityProject
  Users []User
}
type Technology struct {
  Name string
  SVG string
}
type CharityProject struct {
  Name string
  ShortDescription string
  LongDescription string
  CharityName string
  CharityEmail string
  ProjectEmail string
  Location string
  Technologies []Technology
  Archived bool
}
type User struct {
  Username string
  Password string
  Role string
}
func (errorWriter ErrorWriter) Write(p []byte) (n int, err error) { // TODO: fix common code across modules
  color.Red(string(p))
  return 0, nil
}
type ErrorWriter struct {}
var logError *log.Logger

  // flagSet := flag.NewFlagSet("dropDatabase", flag.PanicOnError)
  // err := flagSet.Parse(nil)
  // if (err != nil) { log.Print(err.Error()) }

  // svg, _ := os.ReadFile("../assets/icons/icons8-react.svg")
  // svgString := string(svg)
  // newSvg := stringReplaceFirstGroup("width=\"(\\d+)", svgString, "24")
  // newSvg = stringReplaceFirstGroup("height=\"(\\d+)", newSvg, "24")
  // log.Print(newSvg)
