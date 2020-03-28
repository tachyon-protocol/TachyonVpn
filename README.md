# TachyonVpn
A Decentralized VPN that is secured by Tachyon Protocol and served by our global [node network](https://tachyon.eco/?n=yr8mtzfwee.Network).


## Build


## Usage
### Direct Mode
- For servers which can be accessed from Internet directly (with public IP and public port)
- run server `server`
- run client `client [server's IP]`
### Relay Mode
- For servers which can not be accessed from Internet directly and need another 'Listen Mode' server to relay its traffic
- run relay server `server`
- run exit server `server -UseRelay -RelayServerIp [relay server's IP]`
- run client `client -IsRelay -ServerIp [relay server's IP] -ExitServerClientId [exit server's ClientId]`
### TKey Direct Mode
- run server `server -SelfTKey [server's TKey]`
- run client `client -ServerIp [server's IP] -ServerTKey [server's TKey]`
### TKey Relay Mode
- run relay server `server -SelfTKey [relay server's TKey]`
- run exit server `server -SelfTKey [exit server's TKey] -UseRelay -RelayServerIp [relay server's IP] -RelayServerTKey [relay server's TKey]`
- run client `client -IsRelay -ServerIp [relay server's IP] -ServerTKey [relay server's TKey] -ExitServerClientId [exit server's ClientId] -ExitServerToken [exit server's TKey]`

## Releases
### demo-v6 Support disconnection of client and GC client in server
* Support disconnection of client
* GC client in server
* DHT Demo V1
### demo-v5 Router server is launched to improve stability of VPN connection.
* Router Server
  * VPN server will register itself to Router server
  * VPN client can fetch available IP list from Router server
* VPN Optimization
  * Reduce memory allocation of VPN connection
  * Improve stability of VPN connection
  * Test multiple clients and VPE servers
### demo-v4 Reconnect and Verify Certificate
* Improve security: verify hash of certificate
* Support reconnection between client and server
* Support reconnection between relay server and vpe server
### demo-v3 Support server to verify TKey
Server can verify TKey (e.g. 123456 or vRm4hdY!9cwavRg) in this version. When client tries to connect server with a TKey, server will verify whether TKey is matched or not. 
Server can decide which client has permission to connect with it.
### demo-v3-docker Docker Image Runner of Tachyon Server
To reduce steps of running server on Windows or macOS by Docker, we developed Docker Image Runner for Tachyon Server.
### demo-v2-docker Server runs on Docker for Windows and MacOS
We'll implement native version server for Windows and MacOS in the future.
At the experimental stage, we'll build a Docker image to run Tachyon Server on Windows and MacOS.
The image will be updated with Linux version.
### CheckHelper-v2
### demo-v2 Support P2P Relay Mode
### demo-v1 First Demo
## Tachyon Protocol Plan
### 2020-3-28
* Kademlia Node ID Generation
### 2020-3-21
* DHT IPv6 Tester
* Encrypt Method EXP0
### 2020-3-9
* Improve Test Coverage
* DHT IPv6 Support
### 2020-3-2
* Add server list automatically
* Support 20 Global locations
### 2020-2-17
* DHT FIND_NODE fix nil reference BUG
* Refactor DHT and Encapsulate the message protocol
* Complement all DHT RPC related tests
### 2020-2-10
* Optimize the connection experience
* Optimize the process of adding Servers
### 2020-2-5
* DHT RpcNode API refactor in progress 
### 2020-1-21
* DHT RPC API: Ping
* DHT k-buckets GC
### 2020.1.14
* Refactor k results of DHT RPC API: FIND_NODE/FIND_VALUE
### 2020-1-7
* Improve reliability of DHT RPC's network performance
### 2019-12-31
* Improve test coverage of DHT RPC
* Optimize reliability of routing tables
### 2019-12-24
* DHT store, query, lookup etc.
* DHT V2 implementation in memory
### 2019-12-20
* DHT store, query, lookup etc.
* DHT V2 implementation in memory
### 2019-12-9
* Support disconnection of client
* GC client in server
* Support relay mode in Router server
* DHT Demo V1
### 2019-12-2
* Deploy router server
* automatic RPC generator
* Improve stability of connection
* Optimize performance of memory and bandwidth
* Testing of multiple clients and VPE servers

## Details of demo version
* Router will be a single server for test in this version
    * forward data between clients and servers
    * client and server will not be connected to each other directly in this version
![structure](https://raw.githubusercontent.com/tachyon-protocol/TachyonVpn/master/structure.png)
* Protocol Layers
	* VPN Protocol Layer
		* Packet Type
			* Handshake
			* IpPacket/Traffic
	* Forward Protocol Layer
		* Claim: client or server registers on the Router
		* Forward
	* Encrypt Layer
		* TLS
		* Man-in-the-middle attack: client should not use server's IP, but use server's certificate to identify server
	* Layers Nest
		* when: client connects to server directly:
			* IP Packet > TCP
				* TLS > VPN Protocol > Data IP Packet
		* when: client and server connect to the Router, and the Router will forward traffic
			* IP Packet > TCP
				* TLS > Forward Protocol
				    * TLS > VPN Protocol > Data IP Packet

