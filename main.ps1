$Env:GOOS = "linux"
$Env:GOARCH = "amd64"

mkdir build
mkdir build/lambdas
cd lambdas
  cd lambda1
    go get .
    go build -o ../../build/lambdas/lambda1
  cd ..
  cd users
    go get .
    go build -o ../../build/lambdas/users
  cd ..
  cd charityProjects
    go get .
    go build -o ../../build/lambdas/charityProjects
  cd ..
  cd technologies
    go get .
    go build -o ../../build/lambdas/technologies
  cd ..
  cd login
    go get .
    go build -o ../../build/lambdas/login
  cd ..
  # cd logout
  #   go get .
  #   go build -o ../../build/lambdas/logout
  # cd ..
  cd register
    go get .
    go build -o ../../build/lambdas/register
  cd ..
cd ..

terraform init
terraform plan -out build/terraform-plan
terraform apply build/terraform-plan

