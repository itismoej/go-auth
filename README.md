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
```shell
docker pull mjafari98/go-auth:latest
```
Run the container:
```shell
docker run --rm -p 9090:9090 -p 50051:50051 --env-file=.env -it mjafari98/go-auth:latest
```
Do not forget to modify `--env-file=.env` and replace the `.env` part to the path of
the env file of the project (You may have placed it somewhere else!)

## User Creation (Admin Users Only)
Before we dig into the instructions, RUN the project with the commands in the previous section. 

If you are an ADMIN, you can create users. In order to do that, follow the instructions bellow:
### Create User Using `grpcurl` CLI (gRPC)
1. Copy your *access token* with the command below. Do not forget to replace `ADMIN_USER` & `ADMIN_PASS`.
```shell
grpcurl -plaintext -d '{
  "username":"ADMIN_USER", 
  "password":"ADMIN_PASS"
}' localhost:50051 auth.Auth/Login
```
2. Add authorization header with **Bearer** scheme to the Signup RPC:
```shell
grpcurl -plaintext -d '{
  "username":"new_user", 
  "password":"new_password"
}' -H "Authorization: Bearer jwt" localhost:50051 auth.Auth/Signup
``` 

### Create User Using `curl` CLI (REST)
1. Copy your *access token* with the command below. Do not forget to replace `ADMIN_USER` & `ADMIN_PASS`.
```shell
curl -X POST http://localhost:9090/login -d '{
  "username":"ADMIN_USER", 
  "password":"ADMIN_PASS"
}'
```
2. Add authorization header with **Bearer** scheme to the Signup RPC:
```shell
curl -X POST http://localhost:9090/signup -H "Authorization: Bearer jwt" -d '{
  "username":"new_user", 
  "password":"new_password"
}'
```

## Get List of Users (Admin Users Only)
You can get the list of Users with all information, if you are an ADMIN of course. 

To do that, follow instructions bellow:

### Login
Copy your *access token* with the command below. Do not forget to replace `ADMIN_USER` & `ADMIN_PASS`.
```shell
grpcurl -plaintext -d '{
  "username":"ADMIN_USER", 
  "password":"ADMIN_PASS"
}' localhost:50051 auth.Auth/Login
```

### Ask For User Data 
Add authorization header with **Bearer** scheme to the GetUserInfo RPC and send the ID's in the body:
```shell
grpcurl -plaintext -d '{
  "Id": [11, 2, 31, 4, 9, 120]
}' -H "Authorization: Bearer jwt" localhost:50051 auth.Auth/GetUserInfo
```

This `Id` field can be either an Array, or a single `Id` in integer form:
```shell
grpcurl -plaintext -d '{
  "Id": 11
}' -H "Authorization: Bearer jwt" localhost:50051 auth.Auth/GetUserInfo
```

If the `Id` field is empty, the RPC returns all the users.
```shell
grpcurl -plaintext -d '{
  "Id": []
}' -H "Authorization: Bearer jwt" localhost:50051 auth.Auth/GetUserInfo
```

### All the requests are available in **REST**:
```shell
curl -X POST http://localhost:9090/getusers -d '{
  "Id": [1,2,3]
}' -H "Authorization: Bearer jwt"
```

## Change Password
If you prefer, you can change your password. 

You need your **current** password in body, and a **valid jwt** in the header, to change your password.

### gRPC
You can do it in gRPC call.
```shell
grpcurl -plaintext -d '{
  "oldPassword": "my_old_pass",
  "newPassword": "my_str0ng_new_p4s5word"
}' -H "Authorization: Bearer jwt" localhost:50051 auth.Auth/ChangePassword
```

### REST
You can change it with a POST request.
```shell
curl -X POST http://localhost:9090/user/edit_password -d '{
  "oldPassword": "my_old_pass",
  "newPassword": "my_str0ng_new_p4s5word"
}' -H "Authorization: Bearer jwt"
```

## Refresh The Access Token
The jwt access token has a short life-time, and you should refresh it continuously.

### gRPC
```shell
grpcurl -plaintext -d '{                         
  "token": "your refresh jwt token"
}' localhost:50051 auth.Auth/RefreshAccessToken
```

### REST
```shell
curl -X POST http://localhost:9090/refresh -d '{
  "token": "your refresh jwt token"
}'
```