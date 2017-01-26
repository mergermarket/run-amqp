package runamqp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"log"
)

const entryBody = `<html>
<head>
<title></title>
</head>

<body>
<form action="/entry" method="post">

<input type="text" name="pattern" />
<textarea name="message" ></textarea>

<input type="submit" value="Send" />

</form>

</body>

</html>`

type publisher interface {
	IsReady() bool
	Publish(message []byte, pattern string) error
}

type publisherServer struct {
	router       *http.ServeMux
	publisher    publisher
	exchangeName string
}

func newPublisherServer(publisher publisher) *publisherServer {
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

		fmt.Fprint(w, entryBody)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "POST PLZ", http.StatusMethodNotAllowed)
		return
	}

	log.Println("headers", r.Header)

	r.ParseForm()
	log.Println("form", r.Form)

	var pattern string
	var body []byte

	if len(r.Form) > 0 {
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
	if p.publisher.IsReady() {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Rabbit is up!")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Rabbit did not start up!")
	}
}
