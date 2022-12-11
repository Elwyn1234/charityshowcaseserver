package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"regexp"

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
  if (len(os.Args) < 2) {
    log.Fatal("No subcommands specified. Exiting.")
  }

  createdbFlags := flag.NewFlagSet("dropdb", flag.PanicOnError)
  var createdbFlagsForce bool
  createdbFlags.BoolVar(&createdbFlagsForce, "f", false, "")
  createdbFlags.BoolVar(&createdbFlagsForce, "force", false, "Force the entire database to be dropped and recreated.")

  err := createdbFlags.Parse(os.Args[2:])
  if (err != nil) { log.Print(err.Error()) }

  switch os.Args[1] {
    case "createdb": {
      pool := rootOpenDatabase()
      if (createdbFlagsForce){
        dropDatabase(pool)
      }
      log.Print("Creating charityShowcase Database.")
      _, err := pool.Exec(`CREATE DATABASE charityShowcase;`)
      if (err != nil) { log.Fatal(err.Error()) }
      log.Print("charityShowcase database successfully created!")
      _, err = pool.Exec(`USE charityShowcase;`)
      if (err != nil) { log.Fatal(err.Error()) }
      log.Print("charityShowcase database selected.")
    
      _, err = pool.Exec(`CREATE TABLE charityProject (
        name VARCHAR(50) NOT NULL,
        shortDescription VARCHAR(300) NOT NULL,
        longDescription VARCHAR(5000) NOT NULL,
        PRIMARY KEY (name)
      );`)
      if (err != nil) { log.Fatal(err.Error()) }
      log.Print("charityProject table created!")
    
      _, err = pool.Exec(`CREATE TABLE technology (
        name VARCHAR(32) NOT NULL,
        imageFileName VARCHAR(64) NOT NULL,
        PRIMARY KEY (name)
      )`)
      if (err != nil) { log.Fatal(err.Error()) }
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
      if (err != nil) { log.Fatal(err.Error()) }
      log.Print("technologyToCharityProject table created!")

      _, err = pool.Exec(`CREATE TABLE user (
        username VARCHAR(32) NOT NULL,
        password VARCHAR(32) NOT NULL,
        role ENUM('user', 'editor', 'creator'),
        PRIMARY KEY (username)
      );`)
      if (err != nil) { log.Fatal(err.Error()) }
      log.Print("user table created!")
    }

    case "dropdb": {
      dropDatabase(rootOpenDatabase())
    }
    default: {
      log.Fatal("The subcommand provided is not a recognised subcommand. Exiting.")
    }
  }
}

func rootOpenDatabase() (pool *sql.DB) {
  pool, err := sql.Open("mysql", "root:o1M@2UO4ngwg!i9R$3hvLSVpt@(localhost:3307)/") // TODO: get the password from a file
  if (err != nil) { log.Fatal(err.Error()) }

  pool.SetConnMaxLifetime(0)
  pool.SetMaxIdleConns(3)
  pool.SetMaxOpenConns(3)
  if err := pool.Ping(); err != nil { log.Fatal(err.Error()) }

  return pool
}

func dropDatabase(pool *sql.DB) {
  log.Print("Dropping charityShowcase Database.")
  _, err := pool.Exec("DROP DATABASE charityShowcase")
  if (err != nil) { log.Fatal(err.Error()) }
  log.Print("Database dropped!")
}

  // flagSet := flag.NewFlagSet("dropDatabase", flag.PanicOnError)
  // err := flagSet.Parse(nil)
  // if (err != nil) { log.Print(err.Error()) }

  // svg, _ := os.ReadFile("../assets/icons/icons8-react.svg")
  // svgString := string(svg)
  // newSvg := stringReplaceFirstGroup("width=\"(\\d+)", svgString, "24")
  // newSvg = stringReplaceFirstGroup("height=\"(\\d+)", newSvg, "24")
  // log.Print(newSvg)
