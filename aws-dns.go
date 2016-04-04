package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/go-ini/ini"
	"github.com/miekg/dns"
)

var addresses = make(map[string]string)

const dnsSuffix string = "aws."

var awsRegion = flag.String("region", "ap-southeast-2", "AWS region for API access")
var port = flag.Int("port", 10053, "UDP Port to listen for DNS requests on")
var refreshInterval = flag.Int("refresh", 5, "Number of minutes between refreshing hosts")

func main() {
	flag.Parse()
	go updateAddresses()

	server, addr, _, _ := setupServer()
	dns.HandleFunc(dnsSuffix, awsDNSServer)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer server.Shutdown()
	fmt.Println("Serving on ", addr)
	wg.Wait()
}

func updateAddresses() {
	ticker := time.NewTicker(time.Duration(*refreshInterval) * time.Minute)
	populateAddresses()
	select {
	case <-ticker.C:
		fmt.Println("Updating addresses.")
		populateAddresses()
	}

}

func setupServer() (*dns.Server, string, chan struct{}, error) {
	pc, err := net.ListenPacket("udp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		panic(err)
		// return nil, "", nil, err
	}
	server := &dns.Server{PacketConn: pc, ReadTimeout: time.Hour, WriteTimeout: time.Hour}

	waitLock := sync.Mutex{}
	waitLock.Lock()
	server.NotifyStartedFunc = waitLock.Unlock

	fin := make(chan struct{}, 0)

	go func() {
		server.ActivateAndServe()
		close(fin)
		pc.Close()
	}()

	waitLock.Lock()
	return server, pc.LocalAddr().String(), fin, nil
}

func awsDNSServer(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)

	record := strings.ToLower(req.Question[0].Name)
	fmt.Println("DNS Request: ", record)

	// Lookup the address
	if len(addresses[record]) > 0 {
		m.Extra = make([]dns.RR, 1)
		m.Extra[0] = &dns.A{Hdr: dns.RR_Header{Name: m.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}, A: net.ParseIP(addresses[record])}
		w.WriteMsg(m)
	} else {
		m.SetRcode(req, dns.RcodeNameError)
		m.Authoritative = true

		w.WriteMsg(m)
	}
}

func populateAddresses() {
	for _, profile := range getAvailableAwsProfiles() {
		fmt.Printf("Loading hosts for AWS profile %s\n", strings.ToLower(profile))
		svc := ec2.New(session.New(), &aws.Config{Credentials: credentials.NewSharedCredentials("", strings.ToLower(profile)), Region: aws.String(*awsRegion)})

		// Call the DescribeInstances Operation
		resp, err := svc.DescribeInstances(nil)
		if err != nil {
			fmt.Printf("Unable to load instance details for profile %s: %s\n", profile, err)
			continue
		}

		// resp has all of the response data, pull out instance IDs:
		fmt.Println("> Number of instances: ", len(resp.Reservations))
		for idx := range resp.Reservations {
			for _, inst := range resp.Reservations[idx].Instances {
				name := getNameTagVal(inst.Tags)
				record := strings.ToLower(fmt.Sprintf("%s.%s", *inst.InstanceId, dnsSuffix))
				addresses[record] = *inst.PrivateIpAddress
				fmt.Printf("Added address %s: %s\n", record, *inst.PrivateIpAddress)
				if len(name) != 0 {
					record := strings.ToLower(fmt.Sprintf("%s.%s", parameterizeString(name), dnsSuffix))
					addresses[record] = *inst.PrivateIpAddress
					fmt.Printf("Added address %s: %s\n", record, *inst.PrivateIpAddress)
				}
			}
		}
	}
}

func parameterizeString(input string) string {
	input = strings.ToLower(input)
	re := regexp.MustCompile("[^a-z0-9-_.]+")
	return re.ReplaceAllString(input, "-")
}

func getAvailableAwsProfiles() []string {
	homeDir := os.Getenv("HOME")
	file := filepath.Join(homeDir, ".aws", "credentials")
	config, err := ini.Load(file)
	if err != nil {
		fmt.Printf("Error loading aws credentials file to discover profiles: %s\n", err)
		return []string{"default"}
	}
	return config.SectionStrings()
}

func getNameTagVal(tags []*ec2.Tag) string {
	for _, tag := range tags {
		if *tag.Key == "Name" {
			return *tag.Value
		}
	}

	return ""
}
