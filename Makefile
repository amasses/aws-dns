build: deps
	go build aws-dns.go

deps:
	go get github.com/miekg/dns
	go get github.com/go-ini/ini
	go get github.com/aws/aws-sdk-go/aws
	go get github.com/aws/aws-sdk-go/aws/credentials
	go get github.com/aws/aws-sdk-go/aws/session
	go get github.com/aws/aws-sdk-go/service/ec2

install: build
	sudo mv aws-dns /usr/local/bin/aws-dns
	cp net.amasses.aws-dns.plist ~/Library/LaunchAgents/net.amasses.aws-dns.plist
	sudo cp resolver-config /etc/resolver/aws
	launchctl load -wF ~/Library/LaunchAgents/net.amasses.aws-dns.plist
	@echo "**** Install complete"

uninstall:
	launchctl unload -wF ~/Library/LaunchAgents/net.amasses.aws-dns.plist
	rm -rf ~/Library/LaunchAgents/net.amasses.aws-dns.plist
	sudo rm -rf /etc/resolver/aws
	sudo rm -rf /usr/local/bin/aws-dns
	@echo "**** Uninstall complete"
