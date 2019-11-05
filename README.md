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
  - [Docker](#docker)
  - [Database](#database)  
     * [MySQL](#mysql) 
     * [PostgreSQL](#postgresql) 
     * [BigQuery](#bigquery) 
  - [Datastore](#datastore)
     * [DynamoDB](#dynamodb) 
     * [MongoDB](#mongodb) 
     * [Firestore](#firestore) 
     * [Aerospike](#aerospike) 
  - [Message Bus](#message-bus)
     * [GCP - Pub/Sub](#gcp-pubsub) 
     * [AWS - Simple Queue Service](#aws-simple-queue-service) 
            

## Security

### Private git


The following workflow shows example how to use private git repo without compromising secrets. 

```bash
cd security/git/go/
endly app
```
_Where:_

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



- **git-myaccount** is credentials file for private git repository created by endly -c=git-myaccount

Output:

![Backup Output](/images/secret_git_output.png)


### Database credentials

The following workflow shows example how run database backup without compromising database credentials.


```bash
cd security/database
endly backup.yaml -t=take
endly backup.yaml -t=restore
```

_Where:_

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

The following workflow show how to build a application docker container image.


```bash
cd deplyoment/docker/go
endly app.yaml
```

_Where:_ 

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


![Docker Output](/images/docker_output.png)


Reference: [Endly Docker Service](https://github.com/viant/endly/tree/master/system/docker)


### Developer machine

#### React App

The following workflow shows how to automate react app build and deployment. 


```bash
cd deplyoment/developer/node
endly app.yaml -m
```

_Where:_ 

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

- '-m' option enables interactive mode (endly continues to run unless ctr-c)
- cloud.machine is your localhost or cloud VM
- dev is credentials created for cloud machine to connect with  SSH service, created by ```endly -c=dev```



Output:

![Node Output](/images/node_output.png)


#### Java webapp

The following workflow shows how to automate java webapp build and deployment.


```bash
cd deplyoment/developer/tomcat
endly app.yaml
```

_Where:_ 

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

    
![Tomcat Output](/images/tomcat_output.png)



#### Golang app

The following workflow shows how to automate golang build and deployment.

```bash
cd deplyoment/developer/go
endly app.yaml
```
_Where:_ 

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


![Go Output](/images/go_output.png)

### Hybrid

The following workflow shows how to automate go app build and deployment in hybrid mode.

```bash
cd deplyoment/hybrid/go
endly app.yaml
```

_Where:_ 

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
  buildImage:
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


The following workflow shows how automate cloud functions deployment.

```bash
cd deplyoment/serverless/cloud_functions/go
endly app.yaml
```

_Where:_ 

 
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



![Cloud Function Output](/images/cloud_function_output.png)


References: 
- [Cloud function e2e automation](https://github.com/adrianwit/serverless_e2e/tree/master/cloud_function)
- [Cloud functions endly service](https://github.com/viant/endly/tree/master/system/cloud/gcp/cloudfunctions)

#### Lambda


The following workflow shows how automate lambda deployment.

```bash
cd deplyoment/serverless/lambda/go
endly app.yaml
```

_Where:_ 

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

References:
 - [Lambda e2e automation](https://github.com/adrianwit/serverless_e2e/tree/master/lambda)
-  [Lambda endly service](https://github.com/viant/endly/tree/master/system/cloud/aws/lambda)

## Application State

### Docker

The following workflow shows how automate database/datastore state setup with docekr.

```bash
cd deplyoment/state/docker
endly db
```


_Where:_ 


- [@db.yaml](state/docker/db.yaml)
```yaml
init:
  mydbCredentials: mysql-mydb-root
  mydbSecrets: ${secrets.$mydbCredentials}

pipeline:
  services:
    mysql:
      action: docker:run
      image: mysql:5.7
      name: mydb1
      ports:
        3306: 3306
      env:
        MYSQL_ROOT_PASSWORD: ${mydbSecrets.Password}
    aerospike:
      action: docker:run
      image: 'aerospike/aerospike-server:3.16.0.6'
      name: mydb2
      ports:
        3000: 3000
        3001: 3001
        3002: 3002
        3003: 3003
        8081: 8081
      cmd:
        - asd
        - --config-file
        - /opt/aerospike/etc/aerospike.conf
      entrypoint:
        - /entrypoint.sh
```


![Docker Output](/images/state_docker_output.png)

Reference:

- [Docker endly service](https://github.com/viant/endly/tree/master/system/docker)


### Database

#### Mysql

The following workflow shows how automate MySQL state setup

```bash
cd deplyoment/state/database/mysql
endly setup
```

_Where_

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

- mysql-mydb-root is mysql credential created by ```endly -c=mysql-mydb-root```
- 'mydb/data' is the source folder where *.json data file are matched with database tables.


#### PostgreSQL


The following workflow shows how automate PostgreSQL state setup


```bash
cd deplyoment/state/database/postgresql
endly setup
```

_Where_

- [@setup.yaml](state/database/postgresql/setup.yaml)
```yaml
init:
  mydbCredentials: pq-mydb-root
  mydbSecrets: ${secrets.$mydbCredentials}
  dbIP:
    pg: 127.0.0.1

pipeline:
  services:
    postgresql:
      action: docker:run
      image: postgres:9.6-alpine
      name: mydb
      ports:
        5432: 5432
      env:
        POSTGRES_USER: ${mydbSecrets.Username}
        POSTGRES_PASSWORD: ${mydbSecrets.Password}

  create:
    action: dsunit:init
    datastore: mydb
    config:
      driverName: postgres
      descriptor: host=${dbIP.pg} port=5432 user=[username] password=[password] dbname=[dbname] sslmode=disable
      credentials: $mydbCredentials
    admin:
      datastore: postgres
      ping: true
      config:
        driverName: postgres
        descriptor: host=${dbIP.pg} port=5432 user=[username] password=[password] dbname=postgres sslmode=disable
        credentials: $mydbCredentials
    recreate: true
    scripts:
      - URL: mydb/schema.sql

  load:
    action: dsunit:prepare
    datastore: mydb
    URL: mydb/data
```

- pq-mydb-root is PostgreSQL credential created by ```endly -c=pq-mydb-root```


#### BigQuery

###### Setup

The following workflow shows how automate Big Query state setup

```bash
cd deplyoment/state/database/bigquery/setup
endly setup
```

_Where_ 

[@setup.yaml](state/database/bigquery/setup/setup.yaml)
```yaml
init:
  bqCredentials: gcp-e2e

pipeline:
  create:
    action: dsunit:init
    datastore: mydb
    config:
      driverName: bigquery
      credentials: $bqCredentials
      parameters:
        datasetId: mydb
    scripts:
      - URL: mydb/schema.sql

  load:
    action: dsunit:prepare
    datastore: mydb
    URL: mydb/data
```


###### API copy


The following workflow shows how restore BigQuery data with copy API call.

```bash
cd deplyoment/state/database/bigquery/api
endly copy
```


[@copy.yaml](state/database/bigquery/api/copy.yaml)
```yaml
init:
  i: 0
  gcpCredentials: gcp-e2e
  gcpSecrets: ${secrets.$gcpCredentials}

  src:
    projectID: $gcpSecrets.ProjectID
    datasetID: db1
  dest:
    projectID: $gcpSecrets.ProjectID
    datasetID: db1e2e

pipeline:
  registerSource:
    action: dsunit:register
    datastore: ${src.datasetID}
    config:
      driverName: bigquery
      credentials: $gcpCredentials
      parameters:
        datasetId: $src.datasetID

  readTables:
    action: dsunit:query
    datastore: ${src.datasetID}
    SQL: SELECT table_id AS table FROM `${src.projectID}.${src.datasetID}.__TABLES__`
    post:
      dataset: $Records


  copyTables:
    loop:
      action: print
      message: $i/$Len($dataset) -> $dataset[$i].table

    copyTable:
      action: gcp/bigquery:copy
      logging: false
      credentials: $gcpCredentials
      sourceTable:
        projectID: ${src.projectID}
        datasetID: ${src.datasetID}
        tableID: $dataset[$i].table
      destinationTable:
        projectID: ${dest.projectID}
        datasetID: ${dest.datasetID}
        tableID: $dataset[$i].table

    inc:
      action: nop
      init:
        _ : $i++
    goto:
      when: $i < $Len($dataset)
      action: goto
      task: copyTables

```


### Datastore


#### MongoDB

The following workflow shows how automate MongoDB state setup 


```bash
cd deplyoment/state/datastore/mongo
endly setup
```


- [@setup.yaml](state/datastore/mongo/setup.yaml)
```yaml
pipeline:
  services:
    mongo:
      action: docker:run
      image: mongo:latest
      name: mymongo
      ports:
        27017: 27017

  register:
    action: dsunit:register
    datastore: mydb
    ping: true
    config:
      driverName: mgc
      parameters:
        dbname: mydb
        host: 127.0.0.1
        keyColumn: id

  load:
    action: dsunit:prepare
    datastore: mydb
    URL: mydb/data
```


![Mongo Output](/images/mongo_output.png)


#### Aerospike

The following workflow shows how automate Aerospike state setup

```bash
cd state/datastore/aerospike
endly setup
```


[@setup.yaml](state/datastore/aerospike/setup.yaml)
```yaml
pipeline:
  services:
    aerospike:
      action: docker:run
      image: aerospike/aerospike-server:latest
      name: aero
      ports:
        3000: 3000
        3001: 3001
        3002: 3002
        3004: 3004

  setup:
    create:
      action: dsunit:init
      datastore: aerodb
      ping: true
      config:
        driverName: aerospike
        parameters:
          dbname: aerodb
          excludedColumns: uid
          namespace: test
          host: 127.0.0.1
          port: 3000
          users.keyColumn: uid
      recreate: true

    load:
      action: dsunit:prepare
      datastore: aerodb
      URL: aerodb/data

```

**Setup Data:**
[@users.json](state/datastore/aerospike/aerodb/data/users.json)

```json
[
  {},
  {
    "uid": "${uuid.next}",
    "events": {
      "$AsInt(952319704)": {
        "ttl": 1565478965
      },
      "$AsInt(947840387)": {
        "ttl": 1565479008
      }
    }
  },
  {
    "uid": "${uuid.next}",
    "events": {
      "$AsInt(857513776)": {
        "ttl": 1565479080
      },
      "$AsInt(283419022)": {
        "ttl": 1565479092
      }
    }
  }
]
```

```bash
aql> SELECT * FROM test.users;
+----------------------------------------+---------------------------------------------------------------------+
| PK                                     | events                                                              |
+----------------------------------------+---------------------------------------------------------------------+
| "3b6b7f47-453d-4a07-aff0-879bc85d264c" | MAP('{947840387:{"ttl":1565479008}, 952319704:{"ttl":1565478965}}') |
| "67bf0d31-b9a7-417c-86dd-62c03d2bd60c" | MAP('{283419022:{"ttl":1565479092}, 857513776:{"ttl":1565479080}}') |
+----------------------------------------+---------------------------------------------------------------------+
2 rows in set (0.142 secs)

OK

```

#### DynamoDb

The following workflow shows how automate AWS DynamoDB state setup

```bash
cd state/datastore/dbynamodb
endly setup authWith=myAwsSecret.json
```


[@setup.yaml](state/datastore/dynamo/setup.yaml)

```yaml
init:
  '!mydbCredentials': $params.authWith

pipeline:
  setup:
    action: dsunit:init
    datastore: mydb
    config:
      driverName: dyndb
      credentials: $mydbCredentials
    tables:
      - table: events
        pkColumns:
          - Date
          - EventType
        schemaURL: mydb/schema.json

  load:
    action: dsunit:prepare
    datastore: mydb
    URL: mydb/data
```

![Dynamodb Output](/images/dynamo_output.png)

#### Firestore


### File Storage


### Message Bus

#### AWS Simple Queue Service

#### GCP Pub/Sub




