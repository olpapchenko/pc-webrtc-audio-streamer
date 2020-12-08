package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"com.papchenko.audio.server/audio"
	"com.papchenko.audio.server/rtc"
)

type Session struct {
	Session string `json:"session"`
}

func session(w http.ResponseWriter, r *http.Request) {
	var session Session

	err := json.NewDecoder(r.Body).Decode(&session)

	fmt.Printf("received session descriptor: %s\n", session.Session)

	if err != nil {
		fmt.Printf("%s\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return

	}

	samples := audio.StartCapture()
	serverSessionDesc := rtc.StartWebRtc(samples, session.Session)
	fmt.Printf("sending session descriptor: %s\n", serverSessionDesc)
	w.Write([]byte(serverSessionDesc))
}

func StartHttpServer() {
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/session/", session)

	log.Println("Listening on :3000...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
