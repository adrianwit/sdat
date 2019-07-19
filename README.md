# Software Development Endly Automation Workflow Templates


### Security

- [backup.yaml](security/backup.yaml)
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


### Build and Deployment

#### Developer machine

##### React App

[app.yaml](deplyoment/developer/node/app.yaml)
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


##### Java webapp (tomcat)

[app.yaml](deplyoment/developer/tomcat/app.yaml)
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

  stop:
    action: exec:run
    checkErrors: true
    commands:
      - $tomcatLocation/tomcat/bin/catalina.sh start
      - "echo 'App URL: http://127.0.0.1:8080/app/hello'"
```

```bash
endly app.yaml
```
    
![Tomcat Output](/images/tomcat_output.png)



##### Golang app

[app.yaml](deplyoment/developer/go/app.yaml)
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
      - ls myapp
      - $cmd[1].stdout:/myapp/? rm myapp
      - go build -o myapp

  stop:
    action: process:stop
    input: myapp

  start:
    action: process:start
    directory: $appPath/
    watch: true
    immuneToHangups: true
    command: ./myapp

```
```bash
endly app.yaml
```


![Go Output](/images/go_output.png)


### Application State


#### Database

[setup.yaml](state/database/setup.yaml)
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