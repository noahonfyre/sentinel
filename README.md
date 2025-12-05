# sentinel
Sentinel is a small server application that provides utilities for P2P — especially NAT traversal — built mainly for my personal projects with Go.

## How can I use it?
Get the address of your deployed server and send a UDP packet with the following format to the server:
```
sentinel/<version> #<method>
```
Replace `<version>` with the `sentinelProtocolVersion` variable value in your deployed instance, `<method>` with the information you want from the server.

Below is an example of this format using a deployed instance with protocol version `0`, which triggers the `external_address` and returns the value of this functionality to the sender.
```
sentinel/0 #external_address
```

## Examples
### Python
```python
import socket

target = ("127.0.0.1", 23500)

with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as sock:
    sock.sendto(b"sentinel/0 #external_address", target)

    print(socket.inet_ntoa(sock.recv(4)))
```
This example retrieves the external address of the device and prints it out.

## Endpoints
Currently, sentinel only supports the `external_address` endpoint, which will return the public IP address of the initiator.