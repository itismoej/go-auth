# go-auth
An open-source JWT-Based Authentication System which is developing with golang

# Getting Started

## Dependencies
- You should have Docker installed
- You should place a `.env` file containing base64 encoded public-key & private-key

## Start project
Enter the command below:
```shell script
docker run --rm -p 9090:9090 -p 50051:50051 --env-file=.env -it mjafari98/go-auth:latest
```