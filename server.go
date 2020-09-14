package main

import (
	"calCRUDlator/crudsql"
	"encoding/json"
	"fmt"
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
				fmt.Println(b.sqldriver.MsgJSON)

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
		fmt.Println("Client is done. DONE.")
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

func (b *Broker) createMsg(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	type creating struct {
		UserID string
		Entry  string
	}
	var c creating

	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println(c)
	b.sqldriver.CreateMsg(c.UserID, c.Entry)
	b.sqldriver.ReadMsg()
	b.messages <- string(b.sqldriver.MsgJSON)
}

func (b *Broker) readMsg(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "applications/json")
	b.sqldriver.ReadMsg()
	fmt.Fprintf(w, string(b.sqldriver.MsgJSON))
}

func (b *Broker) updateMsg(w http.ResponseWriter, r *http.Request) {
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
	b.sqldriver.UpdateMsg(c.MessageID, c.UserID, c.Entry)
	log.Println(c.MessageID, c.UserID, c.Entry)
	b.sqldriver.ReadMsg()
	b.messages <- string(b.sqldriver.MsgJSON)
}

func (b *Broker) deleteMsg(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	type creating struct {
		MessageID uint
	}
	var c creating

	err := json.NewDecoder(r.Body).Decode(&c)
	fmt.Println(err)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b.sqldriver.DeleteMsg(c.MessageID)
	log.Println(c)
	b.sqldriver.ReadMsg()
	b.messages <- string(b.sqldriver.MsgJSON)
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

func main() {

	b := &Broker{
		make(map[chan string]bool),
		make(chan (chan string)),
		make(chan (chan string)),
		make(chan string),
		crudsql.Crudsql{},
	}

	b.sqldriver.GetConnection("yourname", "yourpassword", "postgres")

	b.Start()
	http.Handle("/", http.HandlerFunc(handler))
	http.Handle("/connect/", b)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./"))))
	http.HandleFunc("/create/", b.createMsg)
	http.HandleFunc("/read/", b.readMsg)
	http.HandleFunc("/update/", b.updateMsg)
	http.HandleFunc("/delete/", b.deleteMsg)
	//http.ListenAndServe(":8080", nil) // local development
	http.ListenAndServe(":"+os.Getenv("PORT"), nil) // heroku hosting
}
