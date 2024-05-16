# rigelctl

## Build

```
$ go build -o rigelctl main.go
```

## Run

```
./rigelctl --etcd-endpoint localhost:2379,localhost:2380,localhost:2390 --app erp --module hr schema add tmp/sample_schema.json
```

### Sample schema

```
{
    "fields": [
        {
            "name": "host",
            "type": "string",
            "description": "The hostname or IP address of the web server."
        },
        {
            "name": "port",
            "type": "int",
            "description": "The port number on which the web server listens for incoming requests.",
            "constraints": {
                "min": 1,
                "max": 65535
            }
        },
        {
            "name": "enableHttps",
            "type": "bool",
            "description": "Indicates whether HTTPS should be enabled for secure communication."
        }
    ],
    "description": "Configuration schema for a web server application."
}
```

## set a config key
```
./rigelctl --etcd-endpoint localhost:2379,localhost:2380,localhost:2390 --app erp --module hr schema add tmp/sample_schema.json
```

