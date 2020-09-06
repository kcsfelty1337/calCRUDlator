package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

type recentMessages struct {
	Result0 			string		`json:"result0"`
	Result1 			string		`json:"result1"`
	Result2 			string		`json:"result2"`
	Result3 			string		`json:"result3"`
	Result4 			string		`json:"result4"`
	Result5 			string		`json:"result5"`
	Result6 			string		`json:"result6"`
	Result7 			string		`json:"result7"`
	Result8 			string		`json:"result8"`
	Result9 			string		`json:"result9"`
}

type Broker struct {
	clients 			map[chan string]bool
	newClients 			chan chan string
	defunctClients 		chan chan string
	messages 			chan string
	rM recentMessages
}

func (b *Broker) Start() {

	go func() {

		for {

			select {

			case s := <-b.newClients:

				b.clients[s] = true
				log.Println("Added new client")
				res, _ := json.Marshal(b.rM)
				s <- string(res)

			case s := <-b.defunctClients:

				delete(b.clients, s)
				close(s)
				log.Println("Removed client")

			case msg := <-b.messages:

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

	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		b.defunctClients <- messageChan
		log.Println("HTTP connection just closed.")
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	for {
		msg, open := <-messageChan
		if !open {break}

		fmt.Fprintf(w, "data: %s\n\n", msg)

		f.Flush()
	}

	log.Println("Finished HTTP request at ", r.URL.Path)
}

func (b *Broker) clientMsg(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	in, err := ioutil.ReadAll(r.Body)
	if err != nil{
		panic(err)
	}
	b.newMsg(string(in))
	res, _ := json.Marshal(b.rM)
	b.messages <- string(res)
}

func (b *Broker) newMsg (n string){
	b.rM.Result9 = b.rM.Result8
	b.rM.Result8 = b.rM.Result7
	b.rM.Result7 = b.rM.Result6
	b.rM.Result6 = b.rM.Result5
	b.rM.Result5 = b.rM.Result4
	b.rM.Result4 = b.rM.Result3
	b.rM.Result3 = b.rM.Result2
	b.rM.Result2 = b.rM.Result1
	b.rM.Result1 = b.rM.Result0
	b.rM.Result0 = n
	fmt.Println("New eentry"+n)
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

	t.Execute(w, "boss")

	log.Println("Finished HTTP request at", r.URL.Path)
}

func main() {

	b := &Broker{
		make(map[chan string]bool),
		make(chan (chan string)),
		make(chan (chan string)),
		make(chan string),
		recentMessages{
			Result0: "",
			Result1: "",
			Result2: "",
			Result3: "",
			Result4: "",
			Result5: "",
			Result6: "",
			Result7: "",
			Result8: "",
			Result9: "",
		},
	}

	b.Start()
	http.Handle("/connect/", b)
	http.HandleFunc("/clientMsg/", b.clientMsg)
	http.Handle("/", http.HandlerFunc(handler))
	http.ListenAndServe(":5555", nil)
}