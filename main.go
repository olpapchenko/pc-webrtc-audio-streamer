package main

import (
	"com.papchenko.audio.server/audio"
	"com.papchenko.audio.server/rtc"
)

func main() {
	samples := audio.StartCapture()
	rtc.StartWebRtc(samples)
}
