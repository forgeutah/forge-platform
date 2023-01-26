package slack

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

//encore:api public raw method=POST path=/slack/interactive
func InteractiveRouter(w http.ResponseWriter, req *http.Request) {
	// Parse the request body
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	fmt.Print(body)
}
