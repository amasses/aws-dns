build:
	go build aws-dns.go

install: build
	sudo cp aws-dns /usr/local/bin/aws-dns
	cp net.fatzebra.aws-dns.plist ~/Library/LaunchAgents/net.fatzebra.aws-dns.plist
	sudo cp resolver-config /etc/resolver/aws
	launchctl load -wF ~/Library/LaunchAgents/net.fatzebra.aws-dns.plist	
	
