# Sample Golang with Couchbase Web Application

This is a sample RESTful API written with Go and the NoSQL database, Couchbase.

## Downloading the Go Dependencies

To download the Go depedencies for this project, execute the following from the command line:

```
go get github.com/gorilla/mux
go get github.com/satori/go.uuid
go get gopkg.in/couchbase/gocb.v1
```

The `mux` package is a more feature rich option to the already existing `net/http` package for Go. It will be used for managing endpoints and running the server. The `go.uuid` package will allow for UUID values to be created. These unique values will be used as document keys. The `gocb.v1` package is the official Go SDK for Couchbase.

## The Data Model

This application allows for the creation of three different types of NoSQL documents. Each of these documents will represent a customer, a product, or a receipt which is a snapshot of a particular order of products for a customer.

### Customers

```
{
    "id": "12345",
    "type": "customer",
    "firstname": "Nic",
    "lastname": "Raboy",
    "creditcards": [
        {
            "brand": "Visa",
            "number": "123456",
            "expiration": "06-20"
        }
    ]
}
```

### Products

```
{
    "id": "123456",
    "type": "product",
    "name": "Nintendo Switch",
    "price": 299.99
}
```

### Receipts

```
{
    "id": "123456",
    "type": "receipt",
    "customer": {
        "id": "123456",
        "type": "customer",
        "firstname": "Nic",
        "lastname": "Raboy"
    },
    "products": [
        {
            "id": "123456",
            "type": "product",
            "name": "Nintendo Switch",
            "price": 299.99
        }
    ],
    "total": 299.99
}
```

## Running the Application

Before you can run this application, some variables must be changed within the Go code. These variables contain information regarding the Couchbase Server cluster.

```
cluster, _ := gocb.Connect("couchbase://192.168.1.31")
cluster.Authenticate(gocb.PasswordAuthenticator{Username: "demo", Password: "password"})
bucket, _ = cluster.OpenBucket("demo", "")
```

Change the above lines in the **main.go** file to reflect the cluster, bucket, and user information for your setup.

## Resources

[Couchbase Developer Portal](https://developer.couchbase.com)