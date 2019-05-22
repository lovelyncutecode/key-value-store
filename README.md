**Description**

Implement a set of two apps - node and client. 
Node application should act as a storage server which is able to store and retrieve string key-value pairs. In case of multiple nodes launched all nodes should connect to each other and share all keys and values they have including newly added keys on any of the nodes. Node should be able to bootstrap itself with a single IP:port combination of any node already running. 

Client should be able to connect to any node with IP:port and key specified and retrieve value for the corresponding key.
