package main

import (
	dlog "log"
	"time"

	"github.com/pion/webrtc/v4"
)

func main() {

	cfg := config{

		addr: ":8080",
	}
	app := &application{
		config: cfg,
	}
	// Init other state
	trackLocals = map[string]*webrtc.TrackLocalStaticRTP{}
	mux := app.mount()

	// request a keyframe every 3 seconds
	go func() {
		for range time.NewTicker(time.Second * 3).C {
			dispatchKeyFrame()
		}
	}()

	dlog.Fatal(app.run(mux))
}
