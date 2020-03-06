package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"rfc5424"

	"github.com/grandcat/zeroconf"
)

const (
	service  string = "_syslog._udp"
	domain   string = "local"
	waitTime int    = 2
)

var server *zeroconf.ServiceEntry

func searchServer() {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			//			log.Println(entry)
			server = entry
		}
		//		log.Println("No more entries.")
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(waitTime))
	defer cancel()
	err = resolver.Browse(ctx, service, domain, entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	<-ctx.Done()
	// Wait some additional time to see debug messages on go routine shutdown.
	time.Sleep(1 * time.Second)
}

func main() {
	hostname := flag.String("hostname", "aortizhost", "hostname to send")
	appname := flag.String("appname", "app", "app name")
	flag.Parse()
	searchServer()
	if server == nil {
		log.Print("no server found")
		return
	}
	log.Print("found ")
	log.Println(server)
	conn, err := net.Dial("udp", server.AddrIPv4[0].String()+":"+strconv.Itoa(server.Port))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		m := rfc5424.Message{
			Priority:  rfc5424.Daemon | rfc5424.Info,
			Timestamp: time.Now(),
			Hostname:  *hostname,
			AppName:   *appname,
			Message:   []byte(text),
		}
		m.WriteTo(conn)
	}
}
