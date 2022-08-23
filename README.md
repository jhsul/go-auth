## go-auth

Simple, cookie-based authentication server written in Go with Docker and Sqlite.

```bash
docker build -t go-auth github.com/jhsul/go-auth
docker run -it -p 3000:3000 go-auth
```

**Sign Up**

```bash
curl --location --request POST 'http://localhost:3000/signup' \
--header 'Content-Type: application/json' \
--data-raw '{
    "username": "jack",
    "password": "****"
}'

# Welcome, jack
```

**Log In**

```bash
curl --location --request POST 'http://localhost:3000/login' \
--header 'Content-Type: application/json' \
--data-raw '{
    "username": "jack",
    "password": "****"
}'

# Welcome back, jack
```

**Get Authentication Status**

```bash
curl --location --request GET 'localhost:3000/me' \
--header 'Cookie: session_id=****'

# jack
```

**Log Out**

```bash
curl --location --request DELETE 'localhost:3000/me'

# Goodbye, jack
```
