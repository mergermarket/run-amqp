package runamqp

import (
	"net/http"
	"fmt"
	"io/ioutil"
)


type publisher interface {
	IsReady() bool
	Publish(message []byte, pattern string) error
}

type publisherServer struct {
	router *http.ServeMux
	publisher publisher
	exchangeName string
}

func newPublisherServer(publisher publisher) *publisherServer{
	p := new(publisherServer)
	p.publisher = publisher
	p.router = http.NewServeMux()
	p.router.HandleFunc("/entry", p.entry)
	p.router.HandleFunc("/up", p.rabbitup)

	return p
}

func (p *publisherServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

func (p *publisherServer) entry(w http.ResponseWriter, r *http.Request) {

	if !p.publisher.IsReady() {
		http.Error(w, "Cannot publish at the moment, please try later", http.StatusServiceUnavailable)
		return
	}

	if r.Method == http.MethodGet {
		fmt.Fprint(w, "https://github.com/mergermarket/run-amqp/issues/10")
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "POST PLZ", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = p.publisher.Publish(body, r.URL.Query().Get("pattern"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Written body to exchange %s", p.exchangeName)
}

func (p *publisherServer) rabbitup(w http.ResponseWriter, r *http.Request) {
	if p.publisher.IsReady() {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Rabbit is up!")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Rabbit did not start up!")
	}
}