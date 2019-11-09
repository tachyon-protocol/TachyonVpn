# TachyonVpn
## Version 2
* Target Platform：Linux/Darwin/Windows/Mac
* CLI Only
* Features
	* Run as VPN Client
		* search for servers' info from the Router and connect to a server
		* Direct Mode
		* `client [Server's IP]`
		* P2P Relay Mode
		* `client -IsRelay -ServerIp [Relay Server's IP] -ExitClientId [Server's ClientId]`
	* Run as VPN Server
		* start VPN service and register self on the Router
	* And run server as root
		* Listen Mode, for servers which can be accessed from Internet directly (with public IP and public port)
		* `server_linux`
		* Relay Mode, for servers which can not be accessed from Internet directly and need another 'Listen Mode' server to relay its traffic
  * `server -UseRelay -RelayServerIp [Listen Mode Server's IP]`
* Router will be a single server for test in this version
    * forward data between clients and servers
    * client and server will not be connected to each other directly in this version
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
____

  

