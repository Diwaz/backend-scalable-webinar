package main

import "net/http"

func (app *application) checkHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))

}
