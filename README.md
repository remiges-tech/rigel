# Rigel

Remiges Rigel is a product which helps application administrators to manage configuration parameters and their values for one or more live applications.

This repository contains the source code for the Rigel server, client library for Go and the command line interface called rigelctl.

It uses etcd as a backend storage for configuration parameters and their values.

## Installation

`rigelctl` is a cli tool that allows you to interact with Rigel. You can use `rigelctl` to add schemas, set configuration values, retrieve configuration values, and so on.

```
go install github.com/remiges-tech/rigel/cmd/rigelctl@latest

```

It will install `rigelctl` in your `$GOPATH/bin` directory.

Or you can download the latest binray release from https://github.com/remiges-tech/rigel/releases 

## Add a schema


```
rigelctl --etcd-endpoint localhost:2379,localhost:2380,localhost:2390 --app banking_app --module transactions --version 1 schema add banking_schema.json
```


### Sample schema

```
{
    "fields": [
        {
            "name": "api_endpoint",
            "type": "string",
            "description": "The URL endpoint for the banking API."
        },
        {
            "name": "max_transactions_per_day",
            "type": "int",
            "description": "The maximum number of transactions allowed per day.",
            "constraints": {
                "min": 1
            }
        },
        {
            "name": "enable_fraud_detection", 
            "type": "bool",
            "description": "Indicates whether fraud detection should be enabled."
        }
    ],
    "description": "Configuration schema for the banking application's transactions module."
}
```

## set a config key

```
rigelctl --app banking_app --module transactions --version 1 --config prod-us config set api_endpoint "https://api.bankingapp.com"
rigelctl --app banking_app --module transactions --version 1 --config prod-us config set enable_fraud_detection true
rigelctl --app banking_app --module transactions --version 1 --config prod-eu config set enable_fraud_detection true
```

For more details on the available commands and flags, run `rigelctl --help`.


## Usage in Go code


### Usage

Here's an example of how to use the Rigel Go package in your banking application:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/remiges-tech/rigel"
    "github.com/remiges-tech/rigel/etcd"
)

func main() {
    // Create a new EtcdStorage instance
    etcdStorage, err := etcd.NewEtcdStorage([]string{"localhost:2379"})
    if err != nil {
        log.Fatalf("Failed to create EtcdStorage: %v", err)
    }

    // Create a new Rigel instance
    rigelClient := rigel.New(etcdStorage, "banking_app", "transactions", 1, "banking_config")

    // Retrieve configuration values
    apiEndpoint, err := rigelClient.Get(context.Background(), "api_endpoint")
    if err != nil {
        log.Fatalf("Failed to get api_endpoint: %v", err)
    }

    enableFraudDetection, err := rigelClient.GetBool(context.Background(), "enable_fraud_detection")
    if err != nil {
        log.Fatalf("Failed to get enable_fraud_detection: %v", err)
    }

    fmt.Printf("API Endpoint: %s\n", apiEndpoint)
    fmt.Printf("Max Transactions Per Day: %s\n", maxTransactionsPerDay)
    fmt.Printf("Enable Fraud Detection: %s\n", enableFraudDetection)
}
