all:
	go build -ldflags "-s -w" -o icmp-client main.go
	sudo chown root icmp-client
	sudo chmod ugo+s icmp-client
	sudo setcap cap_net_raw+ep icmp-client
clean:
	sudo rm icmp-client