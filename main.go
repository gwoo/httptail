// Copyright 2014 GWoo. All rights reserved.
// The BSD License http://opensource.org/licenses/bsd-license.php.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	addr := flag.String("addr", ":2280", "Addr for server")
	creds := flag.String("creds", "admin:password", "Authentication credentials.")
	dir := flag.String("dir", "/var/log", "Base directory where logs are found.")
	flag.Parse()
	http.HandleFunc("/favicon.ico", http.NotFound)
	auth := Auth{"httptail", *creds}
	http.Handle("/", auth.Handler(TailHandler(*dir, NewBroker())))

	// Find some certs
	_, cerr := os.Open("cert.pem")
	_, kerr := os.Open("key.pem")

	// Serve http if certs not found
	if os.IsNotExist(cerr) || os.IsNotExist(kerr) {
		log.Printf("Serving %s on http://%s\n", *dir, *addr)
		http.ListenAndServe(*addr, nil)
		return
	}
	log.Printf("Serving %s on https://%s\n", *dir, *addr)
	http.ListenAndServeTLS(*addr, "cert.pem", "key.pem", nil)
}

//Tail the request file
func TailHandler(directory string, broker *Broker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI[1:] == "" {
			http.NotFound(w, r)
			return
		}
		file, err := os.Open(fmt.Sprint(directory, "/", r.RequestURI[1:]))
		if err != nil {
			http.NotFound(w, r)
			log.Println(err)
			return
		}
		go tail(file, broker)
		broker.ServeHTTP(w, r)
	})
}

// Tail the file and send messages to Broker
func tail(file *os.File, broker *Broker) {
	_, err := file.Seek(0, os.SEEK_END)
	if err != nil {
		log.Println(err)
		return
	}
	reader := bufio.NewReader(file)
	i := 0
	for {
		if broker.connected <= 0 {
			time.Sleep(time.Second * 1)
			continue
		} else if broker.connected > i {
			broker.Event <- []byte(" ")
		}
		i = len(broker.clients)
		message, err := reader.ReadBytes('\n')
		if err == nil {
			broker.Event <- message
			continue
		}
		if err.Error() != "EOF" {
			broker.Event <- []byte(err.Error())
			return
		}
		time.Sleep(time.Second * 1)
	}
}
