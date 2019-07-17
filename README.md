# Software Development Endly Automation Workflow Templates


### Security

- [backup.yaml](security/backup.yaml)
```yaml
init:
  suffix: $TimeFormat('now', 'yyyyMMdd')
  backupFile: mydb_${sufix}.sql
  dbname: mydb
pipeline:
  take:
    dump:
      action: exec:run
      secrets:
        mydb: root-mydb
      commands:
        - echo 'starting ${dbname} backup'
        - mysqldump -uroot -p ${mydb.Password} ${dbname} > /tmp/$backupFile
    upload:
      action: storage:copy
      source:
        URL: /tmp/mydb.sql
      dest:
        credentials: gs-myservice
        URL: gs://myorgbackup/

  restore:
    download:
      action: storage:copy
      source:
        credentials: gs-myservice
        URL: gs://myorgbackup/$backupFile
      dest:
        URL: /tmp/$backupFile
    load:
      action: exec:run
      secrets:
        mydb: root-mydb
      commands:
        - echo 'starting ${dbname} restore'
        - mysql -uroot -p ${mydb.Password} ${dbname} < /tmp/$backupFile

```

Where

- root-mydb is credentials file (~/.secret/root-mydb.json) created by  ```endly -c=root-mydb```
- gs-myservice is google secrets credential file (~/.secret/root-gs-myservice.json)  created for your service account

```bash
endly backup.yaml -t=task
endly backup.yaml -t=restore
```

Reference: [Endly Secrets](https://github.com/viant/endly/tree/master/doc/secrets)