# Welcome to the @vagasprajr's api project!

This documentation has the goal to explain how to use the API and how to run the project locally.

## How to run the project locally

To run the project locally, you need to have the following tools installed:

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Go](https://golang.org/)

- The project [vagasprajr-mongodb](https://github.com/flaviofrancisco/vagasprajr-mongodb) up and running locally.

### The .env file

First of all, you need to create a `.env` file in the root of the project. You can copy the `.env.example` file and change the values if you want.

Remember to use the connection string of the MongoDB database that you have running locally.

## Install the dependencies

To install the dependencies, you need to run the following command:

```bash
go mod download
```

## Run the project

To run the project, you need to run the following command:

```bash
go run main.go
```


