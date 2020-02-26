# Kratos Test Document

## 1 Explain
### 1.1 Code address
https://github.com/sczj/kratos-selfservice-ui-node/

https://github.com/sczj/kratos-master


## 2 database migration
>My database is MySQL。The database migration command is as follows
```shell
soda migrate up -e development -c ./persistence/sql/.soda.yml
```

`.soda.yml` configuration information,database migration succeeded
```yml
development:
  dialect: "mysql"     
  database: "kratos"    
  host: "localhost"    
  port: "3306"          
  user: "root"          
  password: "admin"     
```
## 3 Compile backend project

### 3.1 Compilation project
>Compile main.go under the project root directory to generate main.exe
```shell
go build main.go
```

### 3.2 Startup project
>Run the following named startup project in the project root directory
```shell
main.exe  serve -c  ./docs/.kratos.yaml 
```
>The project startup profile is as follows,Specific profile path[https://github.com/sczj/kratos-master/blob/master/docs/.kratos.yaml](https://github.com/sczj/kratos-master/blob/master/docs/.kratos.yaml)

```yml
serve:
  admin:
    port: 1234
    host: 127.0.0.1
  public:
    port: 1235
    host: 127.0.0.1

dsn: mysql://root:admin@tcp(localhost:3306)/kratos?parseTime=true&multiStatements=true

selfservice:
  strategies:
    password:
      enabled: true

  logout:
    redirect_to: http://127.0.0.1:3000/auth/login

  login:
    request_lifespan: 10m
    after:
      password:
        -
          job: session
        -
          job: redirect
          config:
            default_redirect_url: http://127.0.0.1:3000/
            allow_user_defined_redirect: true

  registration:
    request_lifespan: 10m
    after:
      password:
        -
          job: session
        -
          job: redirect
          config:
            default_redirect_url: http://127.0.0.1:3000/auth/registration
            allow_user_defined_redirect: true

log:
  level: debug

secrets:
  session:
    - PLEASE-CHANGE-ME-I-AM-VERY-INSECURE

urls:
  login_ui: http://127.0.0.1:3000/auth/login
  registration_ui: http://127.0.0.1:3000/auth/registration
  error_ui: http://127.0.0.1:3000/error
  profile_ui: http://127.0.0.1:3000/auth/profile

  # These are undefined because not available in this demo
  mfa_ui: http://127.0.0.1:3000/
  verify_ui: http://127.0.0.1:3000/

  self:
    public: http://127.0.0.1:1235/  #KRATOS_BROWSER_URL
    admin: http://127.0.0.1:1234/   #KRATOS_ADMIN_URL
  default_return_to: http://127.0.0.1:1235/
  whitelisted_return_to_domains:
    - http://127.0.0.1:1235

hashers:
  argon2:
    parallelism: 1
    memory: 131072
    iterations: 2
    salt_length: 16
    key_length: 16

identity:
  traits:
    default_schema_url: http://127.0.0.1:3000/identity.traits.schema.json

courier:
  smtp:
    connection_uri: smtp://test:test@mailhog:1025/

```
## 4 Start frontend project
startup kratos-selfservice-ui-node
### 4.1 config.ts
>Modify frontend profile `config.ts`,Specific address[https://github.com/sczj/kratos-selfservice-ui-node/blob/master/src/config.ts](https://github.com/sczj/kratos-selfservice-ui-node/blob/master/src/config.ts)



```ts
export default {
  kratos: {
    browser: (process.env.KRATOS_BROWSER_URL || 'http://127.0.0.1:1235').replace(/\/+$/, ''),
    admin: (process.env.KRATOS_ADMIN_URL || 'http://127.0.0.1:1234/').replace(/\/+$/, ''),
    public: (process.env.KRATOS_PUBLIC_URL || 'http://127.0.0.1:1235/').replace(/\/+$/, ''),
  },
  baseUrl: (process.env.BASE_URL || '/').replace(/\/+$/, '') + '/',
  jwksUrl: process.env.JWKS_URL || '/',
  projectName: process.env.PROJECT_NAME || 'SecureApp',
}
```

### 4.2  Compiling front end projects
Execute the following commands in turn

```shell
npm i

npm run start
```

## 5 Test Register
### 5.1  Interface address accessed
```js
#1Visit the login page
http://127.0.0.1:1235/self-service/browser/flows/login

#2Login page address displayed successfully
http://127.0.0.1:3000/auth/login?request=1b447b72-3347-4811-99eb-6cc3c783db7e

#3 Click Register to test
http://127.0.0.1:3000/auth/registration?request=d737a431-dbc0-42f4-a3a9-38293a57e222
# 4 Fill in the account password and other information, click Register
http://127.0.0.1:1235/self-service/browser/flows/registration/strategies/password?request=841b2c5b-d09a-4927-9e20-b70ac379fd8d
# 5 Click Register to jump to the wrong address as follows
{"error":{"code":400,"status":"Bad Request","reason":"CSRF token is missing or invalid.","message":"The request was malformed or contained invalid parameters"}}
```

