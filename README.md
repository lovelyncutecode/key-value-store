**Description**

Implement a set of two apps - node and client. 
Node application should act as a storage server which is able to store and retrieve string key-value pairs. In case of multiple nodes launched all nodes should connect to each other and share all keys and values they have including newly added keys on any of the nodes. Node should be able to bootstrap itself with a single IP:port combination of any node already running. 

Client should be able to connect to any node with IP:port and key specified and retrieve value for the corresponding key.

**Usage**

To run "base" node (node that is not connected with any other node) you should use config file with the following fields:
```
{
  "host": "127.0.0.1",
  "port": 8080,
  "request_timeout": 1
}
```

To run node that is connected to already running node use following structure:
```
{
  "host": "127.0.0.1",
  "port": 8081,
  "node_host": "127.0.0.1",
  "node_port": 8080,
  "request_timeout": 1
}

```

Create and edit config file
```
cp config.json.dist config.json
vi config.json
```

Run node:
```
go run ./cmd/node/main.go -config config.json
```

Client is a CLI app. You can set key value and retrieve it with help of two mods: get-key and set-key.
For set-key mode data should be in the following form: key:value, for get-key mode: key. In 'url' flag specify IP:port of the node you want to connect to.   

Run client:
```
go run ./cmd/client/main.go -url 127.0.0.1:8080  -mode set-key -data key:value
```