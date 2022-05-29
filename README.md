# Criando container

## Criando network

`docker newtwork create inovant`

## Criando containers

`docker-compose up --build -d`

## Rodando api

`docker-compose run --rm -p 8080:8080 api go run main.go`
