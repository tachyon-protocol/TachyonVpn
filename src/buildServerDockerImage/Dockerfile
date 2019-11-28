FROM ubuntu:18.04

WORKDIR /usr/local/bin
RUN apt-get update
RUN apt-get install -y net-tools
RUN apt-get install -y iptables
RUN apt-get install -y iproute2
COPY server ./

CMD ["server"]
