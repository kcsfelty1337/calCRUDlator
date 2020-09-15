package main

import (
	"encoding/json"
	"fmt"
	"github.com/kcsfelty1337/calCRUDlator/crudsql"
	"html/template"
	"log"
	"net/http"
	"os"
)

type Broker struct {
	clients        map[chan string]bool
	newClients     chan chan string
	defunctClients chan chan string
	messages       chan string
	sqldriver      crudsql.Crudsql
}

func (b *Broker) Start() {

	go func() {

		for {

			select {

			case s := <-b.newClients:
				// First put our new client on a list of who to update...
				b.clients[s] = true
				log.Println("Added new client")
				// ...then send them their initial update to populate the results area
				b.sqldriver.ReadMsg()

				s <- string(b.sqldriver.MsgJSON)

			case s := <-b.defunctClients:

				// Take disconnected client off the list
				delete(b.clients, s)
				close(s)
				log.Println("Removed client")

			case msg := <-b.messages:
				// As a new message enters the channel, broadcast it out to each of our clients
				for s := range b.clients {

					s <- msg
				}
				log.Printf("Broadcast message to %d clients", len(b.clients))
			}
		}
	}()
}

func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	messageChan := make(chan string)

	b.newClients <- messageChan

	notify2 := r.Context()
	go func() {
		<-notify2.Done()
		b.defunctClients <- messageChan
		log.Println("Client is done. DONE.")
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	for {
		msg, open := <-messageChan
		if !open {
			break
		}

		fmt.Fprintf(w, "data: %s\n\n", msg)

		f.Flush()
	}

	log.Println("Finished HTTP request at ", r.URL.Path)
}

func handler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	t, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatal("Please remember to have an index.html!")

	}

	t.Execute(w, "Template Completed")

	log.Println("Finished HTTP request at", r.URL.Path)
}

func (b *Broker) entry(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	type creating struct {
		MessageID uint
		UserID    string
		Entry     string
	}
	var c creating

	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "POST": // Create
		b.sqldriver.CreateMsg(c.UserID, c.Entry)

	case "GET": // Read
		// Unused: server will send reads instead of clients requesting them, but this is where
		//     a more robust API could ReadMsg(20) to return 20 most recent entries, etc.
		b.sqldriver.ReadMsg()

	case "PUT": // Update
		b.sqldriver.UpdateMsg(c.MessageID, c.UserID, c.Entry)

	case "DELETE": // Delete
		b.sqldriver.DeleteMsg(c.MessageID)
	}

	b.sqldriver.ReadMsg()
	b.messages <- string(b.sqldriver.MsgJSON)
}

func main() {

	b := &Broker{
		make(map[chan string]bool),
		make(chan (chan string)),
		make(chan (chan string)),
		make(chan string),
		crudsql.Crudsql{},
	}
	conString := os.Getenv("DATABSE_URL")
	b.sqldriver.GetConnection(conString)
	b.Start()
	http.Handle("/", http.HandlerFunc(handler))
	http.Handle("/connect/", b)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./"))))
	http.HandleFunc("/api/entry/", b.entry)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
