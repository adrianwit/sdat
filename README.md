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

```bash
endly backup.yaml -t=task
endly backup.yaml -t=restore
```

Output:

![Backup Output](/images/backup_output.png)

Reference: [Endly Secrets](https://github.com/viant/endly/tree/master/doc/secrets)


### Build and Deployment


### Application State

#### Database