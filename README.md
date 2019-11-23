# TachyonVpn
## Version 3
- Target Platform：Linux/Darwin/Windows/Mac
- CLI Only

## Details of demo version
* Router will be a single server for test in this version
    * forward data between clients and servers
    * client and server will not be connected to each other directly in this version
![structure](https://raw.githubusercontent.com/tachyon-protocol/TachyonVpn/server_token/structure.png)
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
		* when: client and Server connect to the Router, and the Router will forward traffic
			* IP Packet > TCP
				* TLS > Forward Protocol
				    * TLS > VPN Protocol > Data IP Packet

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
