# rigelctl

## Build

```
$ go build -o rigelctl main.go
```

## Run

```
./etcdctl --etcd-endpoint localhost:2379,localhost:2380,localhost:2390 --app erp --module hr schema add tmp/sample_schema.json
```


### Sample schema

```
{
    "name": "webServer",
    "version": 1,
    "fields": [
        {
            "name": "host",
            "type": "string"
        },
        {
            "name": "port",
            "type": "int"
        },
        {
            "name": "enableHttps",
            "type": "bool"
        }
    ],
    "description": "Configuration for a web server application"
}
```

## set a config key
```
./etcdctl --app erp --module hr --version 1 --config test config set host "localhost"
```