# Software Development Endly Automation Workflow Templates

- [Security](#security)
  - [Private Git Repository](#private-git)
  - [Database credentials](#database-credentials)
- [Build And Deployment](#build-and-deployment)
  - [Docker](#docker)
  - [Developer machine](#developer-machine)
     * [React App](#react-app)
     * [Java Tomcat Webapp](#java-webapp)  
     * [Golang app](#golang-app)
  - [Hybrid](#hybrid)
  - [Serverless](#serverless)
     * [Cloud functions](#cloud-functions)
     * [Lambda](#lambda)  
- [Application State](#application-state)
  - [Database](#database)  
     * [MySQL](#mysql) 
     * [PostgreSQL](#postgresql) 
     * [BigQuery](#bigquery) 
     
  - [Datastore](#datastore)
     * [DynamoDD](#dynamodb) 
     * [Firebase](#firebase) 
     * [Aerospike](#aerospike) 
            

## Security

### Private git
- [@app.yaml](security/git/go/app.yaml)
```yaml
init:
  appPath: $Pwd()/myapp

pipeline:

  setTarget:
    action: exec:setTarget
    URL: ssh://127.0.0.1
    credentials: dev

  setSdk:
    action: sdk:set
    sdk: go:1.12

  build:
    action: exec:run
    checkError: true
    terminators:
      - Password
      - Username
    secrets:
      gitSecrets: git-myaccount
    commands:
      - cd $appPath
      - export GIT_TERMINAL_PROMPT=1
      - ls *
      - $cmd[2].stdout:/myapp/? rm myapp
      - export GO111MODULE=on
      - go build -o myapp
      - '$cmd[5].stdout:/Username/? $gitSecrets.username'
      - '$cmd[6].stdout:/Password/? $gitSecrets.password'
      - '$cmd[7].stdout:/Username/? $gitSecrets.username'
      - '$cmd[8].stdout:/Password/? $gitSecrets.password'


  stop:
    action: process:stop
    input: myapp

  start:
    action: process:start
    directory: $appPath/
    env:
      PORT: 8081
    watch: true
    immuneToHangups: true
    command: ./myapp
```

```bash
cd security/git/go/
endly app
```

- Where **git-myaccount** is credentials file for private git repository created by endly -c=git-myaccount

Output:

![Backup Output](/images/secret_git_output.png)


### Database credentials
- [@backup.yaml](security/database/backup.yaml)
```yaml
init:
  suffix: 20190716
  backupFile: mydb_${suffix}.sql
  dbname: mydb
  dbbucket: sdatbackup
  dbIP:
    mysql: 127.0.0.1
pipeline:
  take:
    dump:
      action: exec:run
      systemPaths:
        - /usr/local/mysql/bin
      secrets:
        mydb: mysql-mydb-root
      commands:
        - echo 'starting $dbname backup'
        - mysqldump -uroot -p${mydb.password} -h${dbIP.mysql} $dbname > /tmp/$backupFile
        - history -c
    upload:
      action: storage:copy
      source:
        URL: /tmp/$backupFile
      dest:
        credentials: gcp-myservice
        URL: gs://${dbbucket}/data/

  restore:
    download:
      action: storage:copy
      source:
        credentials: gcp-myservice
        URL: gs://{dbbucket}/data/$backupFile
      dest:
        URL: /tmp/$backupFile
    load:
      action: exec:run
      systemPaths:
        - /usr/local/mysql/bin
      secrets:
        mydb: mysql-mydb-root
      commands:
        - echo 'starting $dbname restore'
        - mysql -uroot -p ${mydb.password} -h${dbIP.mysql} $dbname < /tmp/$backupFile
        - history -c

```

Where

- **mysql-mydb-root** is credentials file (~/.secret/mysql-mydb-root.json) created by  ```endly -c=mysql-mydb-root```
- **gs-myservice** is google secrets credential file (~/.secret/gcp-myservice.json)  created for your service account
- **history -c** clear history for security reason


```bash
cd security/database
endly backup.yaml -t=take
endly backup.yaml -t=restore
```

Output:

![Backup Output](/images/backup_output.png)



Troubleshooting secrets:
To show expanded password set ENDLY_SECRET_REVEAL=true

```bash
export ENDLY_SECRET_REVEAL=true
endly backup.yaml -t=take
```

Reference: [Endly Secrets](https://github.com/viant/endly/tree/master/doc/secrets)


## Build and Deployment



### Docker


- [@app.yaml](deplyoment/docker/go/app.yaml)

```yaml
init:
  buildPath: $Pwd()

pipeline:
  build:
    action: docker:build
    path: ${buildPath}
    nocache: false
    tag:
      image: myapp
      version: '1.0'

  stop:
    action: docker:stop
    images:
      - myapp

  start:
    action: docker:run
    name: myapp
    image: myapp:1.0
    env:
      PORT: 8081
```

```bash
cd deplyoment/docker/go
endly app.yaml
```

![Docker Output](/images/docker_output.png)

Where: 
- [Dockerfile](/deplyoment/docker/go/Dockerfile)
    ```dockerfile
    # transient image
    FROM golang:1.12.7-alpine3.10 as build
    WORKDIR /go/src/app
    COPY myapp .
    ENV GO111MODULE on
    RUN go build -v -o /app
    # final image
    FROM alpine:3.10
    RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
    COPY --from=build /app /app
    CMD ["/app"]
    ```


Reference: [Endly Docker Service](https://github.com/viant/endly/tree/master/system/docker)


### Developer machine

#### React App

- [@app.yaml](deplyoment/developer/node/app.yaml)
```yaml
init:
  sourceCode: $Pwd()/my-app
  appPath: /tmp/my-app

pipeline:
  setTarget:
    action: exec:setTarget
    URL: ssh://cloud.machine
    credentials: dev

  setSdk:
    action: sdk:set
    sdk: node:12

  copy:
    action: storage:copy
    dest: $execTarget
    compress: true
    logging: false
    assets:
      '$sourceCode': '/tmp/my-app'

  build:
    action: exec:run
    checkError: true
    commands:
      - env
      - cd $appPath
      - npm install
      - npm test
  stop:
    action: process:stop
    input: react-scripts/scripts/start.js

  start:
    action: process:start
    directory: $appPath/
    watch: true
    immuneToHangups: true
    command: npm start

```

where
- '-m' option enables interactive mode (endly continues to run unless ctr-c)
- cloud.machine is your localhost or cloud VM
- dev is credentials created for cloud machine to connect with  SSH service, created by ```endly -c=dev```


```bash
cd deplyoment/developer/node
endly app.yaml -m
```

Output:

![Node Output](/images/node_output.png)


#### Java webapp

- [@app.yaml](deplyoment/developer/tomcat/app.yaml)
```yaml
init:
  appPath: $Pwd()/app
  tomcatLocation: /tmp/webapp
  tomcatTarget: $tomcatLocation/tomcat/webapps

pipeline:
  setTarget:
    action: exec:setTarget
    URL: ssh://127.0.0.1
    credentials: dev

  setSdk:
    action: sdk:set
    sdk: jdk:1.8
  setMaven:
    action: deployment:deploy
    appName: maven
    version: 3.5
    baseLocation: /usr/local
  deployTomcat:
    action: deployment:deploy
    appName: tomcat
    version: 7.0
    baseLocation: $tomcatLocation

  build:
    action: exec:run
    checkError: true
    commands:
      - cd $appPath
      - mvn clean package
  deploy:
    action: storage:copy
    source:
      URL: $appPath/target/my-app-1.0.war
    dest:
      URL: $tomcatTarget/app.war
      
  stop:
    action: exec:run
    commands:
      - ps -ef | grep catalina | grep -v grep
      - $cmd[0].stdout:/catalina/ ? $tomcatLocation/tomcat/bin/catalina.sh stop
  start:
    action: exec:run
    checkErrors: true
    commands:
      - $tomcatLocation/tomcat/bin/catalina.sh start
      - "echo 'App URL: http://127.0.0.1:8080/app/hello'"
```

```bash
cd deplyoment/developer/tomcat
endly app.yaml
```
    
![Tomcat Output](/images/tomcat_output.png)



#### Golang app

- [@app.yaml](deplyoment/developer/go/app.yaml)
```yaml
init:
  appPath: $Pwd()/myapp

pipeline:

  setTarget:
    action: exec:setTarget
    URL: ssh://127.0.0.1
    credentials: dev

  setSdk:
    action: sdk:set
    sdk: go:1.12

  build:
    action: exec:run
    checkError: true
    commands:
      - cd $appPath
      - ls *
      - $cmd[1].stdout:/myapp/? rm myapp
      - export GO111MODULE=on 
      - go build -o myapp

  stop:
    action: process:stop
    input: myapp

  start:
    action: process:start
    directory: $appPath/
    env:
      PORT: 8081
    watch: true
    immuneToHangups: true
    command: ./myapp
```

```bash
cd deplyoment/developer/go
endly app.yaml
```

![Go Output](/images/go_output.png)

### Hybrid

- [@app.yaml](deplyoment/hybrid/go/app.yaml)

```yaml
init:
  buildPath: $Pwd()
  appPath: $Pwd()/myapp
pipeline:
  setSdk:
    action: sdk:set
    sdk: go:1.12

  build:
    action: exec:run
    checkError: true
    commands:
      - cd $appPath
      - ls *
      - $cmd[1].stdout:/app/? rm app
      - export GO111MODULE=on
      - export GOOS=linux
      - export CGO=0
      - go build -o app
  build:
    action: docker:build
    path: ${buildPath}
    nocache: false
    tag:
      image: myapp
      version: '1.0'

  start:
    action: exec:run
    systemPaths:
      - /usr/local/bin
    commands:
      - docker-compose down
      - docker-compose up -d
```


```bash
cd deplyoment/hybrid/go
endly app.yaml
```

Where: 
- [Dockerfile](deplyoment/hybrid/go/Dockerfile)
    ```dockerfile
    FROM alpine:3.10
    RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
    COPY myapp/app /app
    CMD ["/app"]
    ```


![Go Output](/images/hybrid_output.png)

### Serverless

#### Cloud functions
 
- [@app.yaml](deplyoment/serverless/cloud_functions/go/app.yaml) 

```yaml
init:
  appPath: $Pwd()/hello
  gcpCredentials: gcp-myservice

pipeline:

  setTarget:
    action: exec:setTarget
    URL: ssh://127.0.0.1
    credentials: dev

  setSdk:
    action: sdk:set
    sdk: go:1.12

  vendor:
    action: exec:run
    commands:
      - unset GOPATH
      - GO111MODULE=on
      - cd ${appPath}
      - go mod vendor

  deploy:
    action: gcp/cloudfunctions:deploy
    credentials: $gcpCredentials
    '@name': HelloWorld
    entryPoint: HelloWorldFn
    runtime: go111
    public: true
    source:
      URL: ${appPath}
```


```bash
cd deplyoment/serverless/go
endly app.yaml
```

![Cloud Function Output](/images/cloud_function_output.png)


Reference: [Cloud function e2e automation](https://github.com/adrianwit/serverless_e2e/tree/master/cloud_function)

#### Lambda

- [@app.yaml](deplyoment/serverless/lambda/go/app.yaml) 

```yaml
init:
  functionRole: lambda-hello
  appPath: $Pwd()/hello
  appArchvive: ${appPath}/app.zip
  awsCredentials: aws-myuser

pipeline:

  setTarget:
    action: exec:setTarget
    URL: ssh://127.0.0.1
    credentials: dev

  setSdk:
    action: sdk:set
    sdk: go:1.12


  deploy:
    build:
      action: exec:run
      checkError: true
      commands:
        - cd ${appPath}
        - unset GOPATH
        - export GOOS=linux
        - export GOARCH=amd64
        - go build -o app
        - zip -j app.zip app

    publish:
      action: aws/lambda:deploy
      credentials: $awsCredentials
      functionname: HelloWorld
      runtime:  go1.x
      handler: app
      code:
        zipfile: $LoadBinary(${appArchvive})
      rolename: $functionRole
      attach:
        - policyarn: arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole

```
![Go Output](/images/lambda_output.png)

Reference: [Cloud function e2e automation](https://github.com/adrianwit/serverless_e2e/tree/master/lambda)


## Application State

### Database

#### Mysql

- [@setup.yaml](state/database/mysql/setup.yaml)
```yaml
init:
  mydbCredentials: mysql-mydb-root
  mydbSecrets: ${secrets.$mydbCredentials}
  dbIP:
    mysql: 127.0.0.1
    
pipeline:
  services:
      mysql:
        action: docker:run
        image: mysql:5.7
        name: dbsync
        ports:
          3306: 3306
        env:
          MYSQL_ROOT_PASSWORD: ${mydbSecrets.Password}

  create:
    mydb:
      action: dsunit:init
      datastore: mydb
      recreate: true
      config:
        driverName: mysql
        descriptor: '[username]:[password]@tcp(${dbIP.mysql}:3306)/[dbname]?parseTime=true'
        credentials: $mydbCredentials
      admin:
        datastore: mysql
        ping: true
        config:
          driverName: mysql
          descriptor: '[username]:[password]@tcp(${dbIP.mysql}:3306)/[dbname]?parseTime=true'
          credentials: $mydbCredentials
      scripts:
        - URL: mydb/schema.sql

  load:
    action: dsunit:prepare
    datastore: mydb
    URL: mydb/data
```

Where
- mysql-mydb-root is mysql credential created by ```endly -c=mysql-mydb-root```

#### PostgreSQL


#### BigQuery


### Datastore

#### DynamoDb

#### Firebase

#### Aerospike


### File Storage

### Message Bus

#### AWS - Simple Queue Service

#### GCP - Pub/Sub


