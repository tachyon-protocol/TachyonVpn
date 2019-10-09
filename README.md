# TachyonVpn
## Version 1
* Target Platform：Linux/Darwin/Windows
* CLI Only
* Features
	* Run as VPN Client
		* search server from Router and connect to server
	* Run as VPN Server
		* listen ports, start VPN service and register on the Router
* Router will be a single server for test in this version
		* forward data between clients and servers
		* clients and servers will not connect to each other in this version
* Protocol Layers
	* VPN Protocol Layer
		* Packet Type
			* Handshake
			* IpPacket/Traffic
	* Forward Protocol Layer
		* Claim: clients or servers register on the Router
		* Forward
	* Encrypt Layer
		* TLS
		* Man-in-the-middle attack: client should not use server's IP but rather server's certificate to identify server, 
	* Layers Nest
		* client connect to server directly:
			* IP Packet > TCP
				* TLS > VPN Protocol > Data IP Packet
		* client and Server connect to router, and router forward traffic
			* IP Packet > TCP
				* TLS > Forward Protocol
				    * TLS > VPN Protocol > Data IP Packet
