package main

import "net/http"

func (app *application) handleWS(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("ok"))
}
