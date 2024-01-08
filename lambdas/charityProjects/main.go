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
      var getArchived bool
      var getNotArchived bool
      var exists bool
      _, exists = request.QueryStringParameters["archived"]
      if exists {
        getArchived = true
      }
      _, exists = request.QueryStringParameters["notArchived"]
      if exists {
        getNotArchived = true
      }

      charityProjects, err := getCharityProjects(getArchived, getNotArchived)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 500
        return response, nil
      }

      responseBodyBytes, err := json.Marshal(charityProjects)
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

      var charityProject CharityProject
      err = json.Unmarshal([]byte(decodedBody), &charityProject)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 400
        return response, nil
      }

      err = postCharityProject(charityProject)
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

      var charityProject CharityProjectUpdate
      err = json.Unmarshal([]byte(decodedBody), &charityProject)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 400
        return response, nil
      }

      err = putCharityProject(charityProject)
      if (err != nil) {
        log.Print(err)
        response.StatusCode = 500
        return response, nil
      }
    }



    case "DELETE": {
      // decodedBody, err := base64.StdEncoding.DecodeString(request.Body)
      // if (err != nil) {
      //   log.Print(err)
      //   response.StatusCode = 400
      //   return response, nil
      // }

      // var charityProject CharityProject
      // err = json.Unmarshal([]byte(decodedBody), &charityProject)
      // if (err != nil) {
      //   log.Print(err)
      //   response.StatusCode = 400
      //   return response, nil
      // }

      // err = deleteCharityProject(charityProject.Name)
      // if (err != nil) {
      //   log.Print(err)
      //   response.StatusCode = 500
      //   return response, nil
      // }
    }



    default: {
      log.Print("Invalid HTTP Method")
      response.StatusCode = 400
      return response, nil
    }
  }



  return response, nil
}






func getCharityProjects(getArchived bool, getNotArchived bool) ([]CharityProject, error) {
  var charityProjects []CharityProject = make([]CharityProject, 0)

  result, err := pool.Query(`SELECT * FROM charityProject`)
  if (err != nil) {
    // logError.Print(err)
    return charityProjects, err
  }

  for result.Next() {
    var charityProject CharityProject;
    err := result.Scan(
        &charityProject.Name,
        &charityProject.ShortDescription,
        &charityProject.LongDescription,
        &charityProject.CharityName,
        &charityProject.CharityEmail,
        &charityProject.ProjectEmail,
        &charityProject.Location,
        &charityProject.Archived)
    if err != nil {
      return charityProjects, err
    }

    if  (getArchived && charityProject.Archived == true) ||
        (getNotArchived && charityProject.Archived == false) ||
        (!getArchived && !getNotArchived) ||
        (getArchived && getNotArchived) {

      charityProjects = append(charityProjects, charityProject)
    }
  }

  for i := 0; i < len(charityProjects); i++ {
    sqlString := fmt.Sprintf(`SELECT technology FROM technologyToCharityProject WHERE charityProject='%v'`, charityProjects[i].Name)
    technologyToCharityProjectResult, err := pool.Query(sqlString)
    if (err != nil) {
      return charityProjects, err
    }
    for technologyToCharityProjectResult.Next() {
      var technologyName string
      if err := technologyToCharityProjectResult.Scan(&technologyName); err != nil {
        return charityProjects, err
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
        return charityProjects, err
      }
      technology.SVG = technologyImageFileName
    }
  }

  log.Print("getCharityProjects: ", charityProjects)

  return charityProjects, nil
}






func postCharityProject(charityProject CharityProject) (error) {
  sqlString := fmt.Sprintf(`INSERT INTO charityProject 
      (name, shortDescription, longDescription, charityName, charityEmail, projectEmail, location)
      VALUES ('%v', '%v', '%v', '%v', '%v', '%v', '%v')`,
      charityProject.Name,
      charityProject.ShortDescription,
      charityProject.LongDescription,
      charityProject.CharityName,
      charityProject.CharityEmail,
      charityProject.ProjectEmail,
      charityProject.Location)
  log.Print("createCharityProject: ", sqlString)
  _, err := pool.Exec(sqlString)
  if (err != nil) {
    return err
  }

  for i := 0; i < len(charityProject.Technologies); i++ {
    sqlString = fmt.Sprintf(`INSERT INTO technologyToCharityProject
        (technology, charityProject)
        VALUES ('%v', '%v')`,
        charityProject.Technologies[i].Name,
        charityProject.Name)
    log.Print("createCharityProject: ", sqlString)
    _, err = pool.Exec(sqlString)
    if (err != nil) {
      return err
    }
  }

  return nil
}






func putCharityProject(charityProject CharityProjectUpdate) (error) {
  _, err := pool.Exec(`UPDATE charityProject SET
      name=IF(?='', name, ?),
      shortDescription=IF(?='', shortDescription, ?),
      longDescription=IF(?='', longDescription, ?),
      charityName=IF(?='', charityName, ?),
      charityEmail=IF(?='', charityEmail, ?),
      projectEmail=IF(?='', projectEmail, ?),
      location=IF(?='', location, ?),
      archived=? WHERE name=?`, // TODO: what if archived is not set in the json payload
      charityProject.Name,
      charityProject.Name,
  
      charityProject.ShortDescription,
      charityProject.ShortDescription,
  
      charityProject.LongDescription,
      charityProject.LongDescription,
  
      charityProject.CharityName,
      charityProject.CharityName,
  
      charityProject.CharityEmail,
      charityProject.CharityEmail,
  
      charityProject.ProjectEmail,
      charityProject.ProjectEmail,
  
      charityProject.Location,
      charityProject.Location,
  
      charityProject.Archived,
      charityProject.OldName) // TODO: this sql driver doesn't support named parameters, is there one that does. so that we can replace the above with the below

  // _, err = pool.Exec(`UPDATE charityProject SET name=IF(@name='', name, @name), shortDescription=IF(@shortDescription='', shortDescription, @shortDescription), longDescription=IF(@longDescription='', longDescription, @longDescription), archived=IF(@archived='', archived, @archived) WHERE name=@oldName`,
  //   sql.Named("name", charityProject.Name),
  //   sql.Named("shortDescription", charityProject.ShortDescription),
  //   sql.Named("longDescription", charityProject.LongDescription),
  //   sql.Named("archived", charityProject.Archived),
  //   sql.Named("oldName", charityProject.OldName))

  if (err != nil) {
    return err
  }

  for i := 0; i < len(charityProject.Technologies); i++ {
    sqlString := fmt.Sprintf(`UPDATE technologyToCharityProject
        SET technology='%v'
        WHERE technology='%v'
        AND charityProject='%v'`,
        charityProject.Technologies[i].Name,
        charityProject.Technologies[i].OldName,
        charityProject.Name)
    log.Print("updateCharityProject: ", sqlString)
    _, err = pool.Exec(sqlString)
    if (err != nil) {
      return err
    }
  }

  return nil
}











var pool *sql.DB
var privateKey *rsa.PrivateKey

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
  CharityName string
  CharityEmail string
  ProjectEmail string
  Location string
  Technologies []TechnologyUpdate
  Archived bool
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
