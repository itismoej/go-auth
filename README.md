# go-auth
An open-source JWT-Based Authentication System which is developing with golang

# Getting Started

## Dependencies
- You should have Docker installed
- You should place a `.env` file containing base64 encoded public-key & private-key; And there should be two variables for the super-user creation named `ADMIN_USER` & `ADMIN_PASS`. The env file should be like below.
```
PUBLIC_KEY=abase64encodedpublickey
PRIVATE_KEY=abase64encodedprivatekey
ADMIN_USER=some_user
ADMIN_PASS=Some!Str0ng_p4ss
```
- Do **NOT** forget to change credentials of the Admin user 
- The keys **SHOULD** be Elliptic Curve Digital Signature Algorithm (ECDSA), and they **SHOULD** be encoded to base64 format

## Start project
Pull the project image:
```shell script
docker pull mjafari98/go-auth:latest
```
Run the container:
```shell script
docker run --rm -p 9090:9090 -p 50051:50051 --env-file=.env -it mjafari98/go-auth:latest
```
Do not forget to modify `--env-file=.env` and replace the `.env` part to the path of
the env file of the project (You may have placed it somewhere else!)

## User Creation (Admin Users Only)
Before we dig into the instructions, RUN the project with the commands in the previous section. 

If you are an ADMIN, you can create users. In order to do that, follow the instructions bellow:
### Create User Using `grpcurl` CLI (gRPC)
1. Copy your *access token* with the command below. Do not forget to replace `ADMIN_USER` & `ADMIN_PASS`.
```shell script
grpcurl -plaintext -d '{
  "username":"ADMIN_USER", 
  "password":"ADMIN_PASS"
}' localhost:50051 auth.Auth/Login
```
2. Add authorization header with **Bearer** scheme to the Signup RPC:
```shell script
grpcurl -plaintext -d '{
  "username":"new_user", 
  "password":"new_password"
}' -H "Authorization: Bearer jwttokenhere" localhost:50051 auth.Auth/Signup
``` 

### Create User Using `curl` CLI (REST)
1. Copy your *access token* with the command below. Do not forget to replace `ADMIN_USER` & `ADMIN_PASS`.
```shell script
curl -X POST http://localhost:9090/login -d '{
  "username":"ADMIN_USER", 
  "password":"ADMIN_PASS"
}'
```
2. Add authorization header with **Bearer** scheme to the Signup RPC:
```shell script
curl -X POST http://localhost:9090/signup -H "Authorization: Bearer jwttokenhere" -d '{
  "username":"new_user", 
  "password":"new_password"
}'
``` 
