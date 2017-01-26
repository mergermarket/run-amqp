package runamqp

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const entryBody = `<html>
<head>
<title>Helper form</title>
<style type="text/css">
.form-style-5{
    max-width: 500px;
    padding: 10px 20px;
    background: #f4f7f8;
    margin: 10px auto;
    padding: 20px;
    background: #f4f7f8;
    border-radius: 8px;
    font-family: Georgia, "Times New Roman", Times, serif;
}
.form-style-5 fieldset{
    border: none;
}
.form-style-5 legend {
    font-size: 1.4em;
    margin-bottom: 10px;
}
.form-style-5 label {
    display: block;
    margin-bottom: 8px;
}
.form-style-5 input[type="text"],
.form-style-5 input[type="date"],
.form-style-5 input[type="datetime"],
.form-style-5 input[type="email"],
.form-style-5 input[type="number"],
.form-style-5 input[type="search"],
.form-style-5 input[type="time"],
.form-style-5 input[type="url"],
.form-style-5 textarea,
.form-style-5 select {
    font-family: Georgia, "Times New Roman", Times, serif;
    background: rgba(255,255,255,.1);
    border: none;
    border-radius: 4px;
    font-size: 16px;
    margin: 0;
    outline: 0;
    padding: 7px;
    width: 100%;
    box-sizing: border-box;
    -webkit-box-sizing: border-box;
    -moz-box-sizing: border-box;
    background-color: #e8eeef;
    color:#8a97a0;
    -webkit-box-shadow: 0 1px 0 rgba(0,0,0,0.03) inset;
    box-shadow: 0 1px 0 rgba(0,0,0,0.03) inset;
    margin-bottom: 30px;

}
.form-style-5 input[type="text"]:focus,
.form-style-5 input[type="date"]:focus,
.form-style-5 input[type="datetime"]:focus,
.form-style-5 input[type="email"]:focus,
.form-style-5 input[type="number"]:focus,
.form-style-5 input[type="search"]:focus,
.form-style-5 input[type="time"]:focus,
.form-style-5 input[type="url"]:focus,
.form-style-5 textarea:focus,
.form-style-5 select:focus{
    background: #d2d9dd;
}
.form-style-5 select{
    -webkit-appearance: menulist-button;
    height:35px;
}
.form-style-5 .number {
    background: #1abc9c;
    color: #fff;
    height: 30px;
    width: 30px;
    display: inline-block;
    font-size: 0.8em;
    margin-right: 4px;
    line-height: 30px;
    text-align: center;
    text-shadow: 0 1px 0 rgba(255,255,255,0.2);
    border-radius: 15px 15px 15px 0px;
}

.form-style-5 input[type="submit"],
.form-style-5 input[type="button"]
{
    position: relative;
    display: block;
    padding: 19px 39px 18px 39px;
    color: #FFF;
    margin: 0 auto;
    background: #1abc9c;
    font-size: 18px;
    text-align: center;
    font-style: normal;
    width: 100%;
    border: 1px solid #16a085;
    border-width: 1px 1px 3px;
    margin-bottom: 10px;
}
.form-style-5 input[type="submit"]:hover,
.form-style-5 input[type="button"]:hover
{
    background: #109177;
}
</style>
</head>
<body>
<div class="form-style-5">
<form method="post" action="/entry">
<fieldset>
<legend><span class="number">1</span> Pattern</legend>
<input type="text" name="pattern" placeholder="Routing key pattern">
</select>
</fieldset>
<fieldset>
<legend><span class="number">2</span> Message</legend>
<textarea rows="10" name="message" placeholder="Your message to publish to EXCHANE_NAME exchange"></textarea>
</fieldset>
<input type="submit" value="Send" />
</form>
</div>
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

func newPublisherServer(publisher publisher, exchangeName string) *publisherServer {
	p := new(publisherServer)
	p.publisher = publisher
	p.router = http.NewServeMux()
	p.exchangeName = exchangeName
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
	if p.publisher.IsReady() {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Rabbit is up!")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Rabbit did not start up!")
	}
}
