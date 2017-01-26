package runamqp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"html/template"
)

type publisher interface {
	IsReady() bool
	Publish(message []byte, pattern string) error
}

type publisherServer struct {
	router       *http.ServeMux
	publisher    publisher
	exchangeName string
	form *template.Template
	logger logger
}

func newPublisherServer(publisher publisher, exchangeName string, logger logger) *publisherServer {
	p := new(publisherServer)

	p.publisher = publisher
	p.logger = logger

	p.router = http.NewServeMux()
	p.exchangeName = exchangeName
	p.router.HandleFunc("/entry", p.entry)
	p.router.HandleFunc("/up", p.rabbitup)

	t, err := template.ParseFiles("form.html")

	if err != nil {
		p.logger.Error("Problem parsing form.html template", err)
	}

	p.form = t

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
		p.form.Execute(w, nil)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "POST PLZ", http.StatusMethodNotAllowed)
		return
	}

	var pattern string
	var body []byte

	if contentTypes, ok := r.Header["Content-Type"]; ok && contentTypes[0] == "application/x-www-form-urlencoded" {

		err := r.ParseForm()

		if err != nil {
			http.Error(w, "Could not get the Form data", http.StatusServiceUnavailable)
			return
		}

		pattern = r.Form.Get("pattern")
		body = []byte(r.Form.Get("message"))
	} else {

		defer r.Body.Close()

		b, err := ioutil.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		body = b
		pattern = r.URL.Query().Get("pattern")

	}

	err := p.publisher.Publish(body, pattern)

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
		fmt.Fprintf(w, "Rabbit is up!")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Rabbit did not start up!")
	}
}
