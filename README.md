# TachyonVpn
## Version 1 Draft
* Target Platform：Linux/Darwin/Windows
* CLI Only
* Features
	* Run as VPN Client
		* search for servers' info from the Router and connect to a server
	* Run as VPN Server
		* start VPN service and register self on the Router
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
