// Copyright 2014 GWoo. All rights reserved.
// The BSD License http://opensource.org/licenses/bsd-license.php.
// Based on https://gist.github.com/ismasan/3fb75381cd2deb6bfa9c
package main

import (
	"fmt"
	"log"
	"net/http"
)

type Broker struct {

	// Events are pushed to this channel by the main events-gathering routine
	Event chan []byte

	// the number of connected clients
	connected int

	// New client connections
	joined chan chan []byte

	// Closed client connections
	exited chan chan []byte

	// Client connections registry
	clients map[chan []byte]bool
}

// Instantiate a broker and listen for events
func NewBroker() *Broker {
	broker := &Broker{
		Event:   make(chan []byte, 1),
		joined:  make(chan chan []byte),
		exited:  make(chan chan []byte),
		clients: make(map[chan []byte]bool),
	}
	go broker.listen()
	return broker
}

// http handle interface
func (broker *Broker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	// Make sure that the writer supports flushing.
	//
	flusher, ok := rw.(http.Flusher)

	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Broker's connections registry
	client := make(chan []byte)

	// Signal the broker that we have a new connection
	broker.joined <- client

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		broker.exited <- client
	}()

	// Listen to connection close and un-register messageChan
	notify := rw.(http.CloseNotifier).CloseNotify()

	go func() {
		<-notify
		broker.exited <- client
	}()

	for {

		// Write to the ResponseWriter
		// Server Sent Events compatible
		fmt.Fprintf(rw, "%s", <-client)

		// Flush the data immediatly instead of buffering it for later.
		flusher.Flush()
	}

}

// Listen for events
func (broker *Broker) listen() {
	for {
		select {
		case s := <-broker.joined:

			// A new client has connected.
			// Register their message channel
			broker.clients[s] = true
			broker.connected = len(broker.clients)
			log.Printf("Client added. %d registered clients", len(broker.clients))
		case s := <-broker.exited:

			// A client has dettached and we want to
			// stop sending them messages.
			delete(broker.clients, s)
			broker.connected = len(broker.clients)
			log.Printf("Removed client. %d registered clients", len(broker.clients))
		case event := <-broker.Event:

			// We got a new event from the outside!
			// Send event to all connected clients
			for client, _ := range broker.clients {
				client <- event
			}
		}
	}

}
