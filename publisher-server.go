package runamqp

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
)

type publisher interface {
	IsReady() bool
	Publish(message []byte, options *PublishOptions) error
}

type viewModel struct {
	ExchangeName string
}

type publisherServer struct {
	router       *http.ServeMux
	publisher    publisher
	exchangeName string
	entryForm    *template.Template
	logger       logger
	viewModel    viewModel
}

func newPublisherServer(publisher publisher, exchangeName string, logger logger) *publisherServer {
	p := new(publisherServer)

	p.publisher = publisher
	p.logger = logger

	p.router = http.NewServeMux()
	p.exchangeName = exchangeName
	p.router.HandleFunc("/entry", p.entry)
	p.router.HandleFunc("/up", p.rabbitup)

	entryForm, err := template.ParseFiles("entryForm.html")

	if err != nil {
		p.logger.Error("Problem parsing entryForm.html template", err)
	}

	p.entryForm = entryForm
	p.viewModel = viewModel{
		ExchangeName: exchangeName,
	}

	return p
}

func (p *publisherServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

func (p *publisherServer) entry(w http.ResponseWriter, r *http.Request) {
	p.logger.Debug(p.exchangeName, "Entry form hit", r.Method)

	if !p.publisher.IsReady() {
		http.Error(w, "Cannot publish at the moment, please try later", http.StatusServiceUnavailable)
		return
	}

	if r.Method == http.MethodGet {

		if err := p.entryForm.Execute(w, p.viewModel); err != nil {
			http.Error(w, fmt.Sprintf("Problem rendering templae %v", err), http.StatusInternalServerError)
		}
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "POST PLZ", http.StatusMethodNotAllowed)
		return
	}

	var pattern string
	var body []byte
	var priority uint8
	var publishToQueue string

	if contentTypes, ok := r.Header["Content-Type"]; ok && contentTypes[0] == "application/x-www-form-urlencoded" {

		err := r.ParseForm()

		if err != nil {
			http.Error(w, "Could not get the Form data", http.StatusServiceUnavailable)
			return
		}

		pattern = r.Form.Get("pattern")
		body = []byte(r.Form.Get("message"))
		priorityUint64, _ := strconv.ParseUint(r.Form.Get("priority"), 10, 8)
		priority = uint8(priorityUint64)
		publishToQueue = r.Form.Get("publishToQueue")

	} else {

		defer r.Body.Close()

		b, err := ioutil.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		body = b
		pattern = r.URL.Query().Get("pattern")
		priorityUint64, _ := strconv.ParseUint(r.URL.Query().Get("priority"), 10, 8)
		priority = uint8(priorityUint64)
		publishToQueue = r.URL.Query().Get("publishToQueue")
	}

	options := &PublishOptions{Priority: priority, Pattern: pattern, PublishToQueue: publishToQueue}

	err := p.publisher.Publish(body, options)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Written body to exchange %s", p.exchangeName)
}

func (p *publisherServer) rabbitup(w http.ResponseWriter, r *http.Request) {
	p.logger.Debug(p.exchangeName, "Rabbit up hit")
	if p.publisher.IsReady() {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Rabbit is up!")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "Rabbit did not start up!")
	}
}
