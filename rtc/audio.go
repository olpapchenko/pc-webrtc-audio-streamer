package rtc

import (
	"context"
	"fmt"
	"time"

	"com.papchenko.audio.server/audio"
	"com.papchenko.audio.server/utils"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

//StartWebRtc - initializes webRTC engine, ready to get opus samples
func StartWebRtc(sampleData <-chan audio.SampleData) {

	// Create a new RTCPeerConnection
	m := webrtc.MediaEngine{}

	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 2, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(&m))

	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{},
			},
		},
	})

	if err != nil {
		panic(err)
	}
	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())

	// Create a audio track
	audioTrack, audioTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion")
	if audioTrackErr != nil {
		panic(audioTrackErr)
	}
	if _, audioTrackErr = peerConnection.AddTrack(audioTrack); audioTrackErr != nil {
		panic(audioTrackErr)
	}

	go func() {

		// Wait for connection established
		<-iceConnectedCtx.Done()

		for {

			data := <-sampleData
			sampleDuration, parseErr := time.ParseDuration(fmt.Sprintf("%dms", data.N))
			if parseErr != nil {
				panic(parseErr)
			}

			if oggErr := audioTrack.WriteSample(media.Sample{Data: data.Samples, Duration: sampleDuration}); oggErr != nil {
				panic(oggErr)
			}

			time.Sleep(sampleDuration)
		}
	}()

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		}
	})

	// Wait for the offer to be pasted
	offer := webrtc.SessionDescription{}
	utils.Decode("eyJ0eXBlIjoib2ZmZXIiLCJzZHAiOiJ2PTBcclxubz0tIDQzMTA1NjgyMDc2ODk2Mzk2NzkgMiBJTiBJUDQgMTI3LjAuMC4xXHJcbnM9LVxyXG50PTAgMFxyXG5hPWdyb3VwOkJVTkRMRSAwIDFcclxuYT1tc2lkLXNlbWFudGljOiBXTVNcclxubT12aWRlbyAzNzQzOSBVRFAvVExTL1JUUC9TQVZQRiA5NiA5NyA5OCA5OSAxMDIgMTI1IDEwNFxyXG5jPUlOIElQNCAxOTIuMTY4LjEuMTJcclxuYT1ydGNwOjkgSU4gSVA0IDAuMC4wLjBcclxuYT1jYW5kaWRhdGU6MTQ5MzM5OTEzOSAxIHVkcCAyMTEzOTM3MTUxIDE5Mi4xNjguMS4xMiAzNzQzOSB0eXAgaG9zdCBnZW5lcmF0aW9uIDAgbmV0d29yay1jb3N0IDk5OVxyXG5hPWljZS11ZnJhZzo5NnNBXHJcbmE9aWNlLXB3ZDo3dnFkMmUydjBVc0ZxbU43ejc3ZUQ5SjVcclxuYT1pY2Utb3B0aW9uczp0cmlja2xlXHJcbmE9ZmluZ2VycHJpbnQ6c2hhLTI1NiA2NDpERDo5ODpGNToyNToyRToxNDpFRDoxOToxQTpGNjo4ODo2RDpFMjo0RTpDOTo4RTo4RDozNzo2ODo4ODo0QzpDMTpEODpCMTo1OTpEQjozNjo1Qzo2OTpFNDpGN1xyXG5hPXNldHVwOmFjdHBhc3NcclxuYT1taWQ6MFxyXG5hPWV4dG1hcDoxIHVybjppZXRmOnBhcmFtczpydHAtaGRyZXh0OnRvZmZzZXRcclxuYT1leHRtYXA6MiBodHRwOi8vd3d3LndlYnJ0Yy5vcmcvZXhwZXJpbWVudHMvcnRwLWhkcmV4dC9hYnMtc2VuZC10aW1lXHJcbmE9ZXh0bWFwOjMgdXJuOjNncHA6dmlkZW8tb3JpZW50YXRpb25cclxuYT1leHRtYXA6NCBodHRwOi8vd3d3LmlldGYub3JnL2lkL2RyYWZ0LWhvbG1lci1ybWNhdC10cmFuc3BvcnQtd2lkZS1jYy1leHRlbnNpb25zLTAxXHJcbmE9ZXh0bWFwOjUgaHR0cDovL3d3dy53ZWJydGMub3JnL2V4cGVyaW1lbnRzL3J0cC1oZHJleHQvcGxheW91dC1kZWxheVxyXG5hPWV4dG1hcDo2IGh0dHA6Ly93d3cud2VicnRjLm9yZy9leHBlcmltZW50cy9ydHAtaGRyZXh0L3ZpZGVvLWNvbnRlbnQtdHlwZVxyXG5hPWV4dG1hcDo3IGh0dHA6Ly93d3cud2VicnRjLm9yZy9leHBlcmltZW50cy9ydHAtaGRyZXh0L3ZpZGVvLXRpbWluZ1xyXG5hPWV4dG1hcDo4IGh0dHA6Ly93d3cud2VicnRjLm9yZy9leHBlcmltZW50cy9ydHAtaGRyZXh0L2NvbG9yLXNwYWNlXHJcbmE9ZXh0bWFwOjkgdXJuOmlldGY6cGFyYW1zOnJ0cC1oZHJleHQ6c2RlczptaWRcclxuYT1leHRtYXA6MTAgdXJuOmlldGY6cGFyYW1zOnJ0cC1oZHJleHQ6c2RlczpydHAtc3RyZWFtLWlkXHJcbmE9ZXh0bWFwOjExIHVybjppZXRmOnBhcmFtczpydHAtaGRyZXh0OnNkZXM6cmVwYWlyZWQtcnRwLXN0cmVhbS1pZFxyXG5hPXNlbmRyZWN2XHJcbmE9bXNpZDotIDRkNzlhMzkxLTUzNTEtNGZkOS1iN2Y4LTA4MTY3ZWIzYTExMFxyXG5hPXJ0Y3AtbXV4XHJcbmE9cnRjcC1yc2l6ZVxyXG5hPXJ0cG1hcDo5NiBWUDgvOTAwMDBcclxuYT1ydGNwLWZiOjk2IGdvb2ctcmVtYlxyXG5hPXJ0Y3AtZmI6OTYgdHJhbnNwb3J0LWNjXHJcbmE9cnRjcC1mYjo5NiBjY20gZmlyXHJcbmE9cnRjcC1mYjo5NiBuYWNrXHJcbmE9cnRjcC1mYjo5NiBuYWNrIHBsaVxyXG5hPXJ0cG1hcDo5NyBydHgvOTAwMDBcclxuYT1mbXRwOjk3IGFwdD05NlxyXG5hPXJ0cG1hcDo5OCBWUDkvOTAwMDBcclxuYT1ydGNwLWZiOjk4IGdvb2ctcmVtYlxyXG5hPXJ0Y3AtZmI6OTggdHJhbnNwb3J0LWNjXHJcbmE9cnRjcC1mYjo5OCBjY20gZmlyXHJcbmE9cnRjcC1mYjo5OCBuYWNrXHJcbmE9cnRjcC1mYjo5OCBuYWNrIHBsaVxyXG5hPWZtdHA6OTggcHJvZmlsZS1pZD0wXHJcbmE9cnRwbWFwOjk5IHJ0eC85MDAwMFxyXG5hPWZtdHA6OTkgYXB0PTk4XHJcbmE9cnRwbWFwOjEwMiByZWQvOTAwMDBcclxuYT1ydHBtYXA6MTI1IHJ0eC85MDAwMFxyXG5hPWZtdHA6MTI1IGFwdD0xMDJcclxuYT1ydHBtYXA6MTA0IHVscGZlYy85MDAwMFxyXG5hPXNzcmMtZ3JvdXA6RklEIDM3MzU0MTY4MDYgNDY4Mzc1NThcclxuYT1zc3JjOjM3MzU0MTY4MDYgY25hbWU6V0NBWmtmcEV4WGxhQWxaK1xyXG5hPXNzcmM6MzczNTQxNjgwNiBtc2lkOi0gNGQ3OWEzOTEtNTM1MS00ZmQ5LWI3ZjgtMDgxNjdlYjNhMTEwXHJcbmE9c3NyYzozNzM1NDE2ODA2IG1zbGFiZWw6LVxyXG5hPXNzcmM6MzczNTQxNjgwNiBsYWJlbDo0ZDc5YTM5MS01MzUxLTRmZDktYjdmOC0wODE2N2ViM2ExMTBcclxuYT1zc3JjOjQ2ODM3NTU4IGNuYW1lOldDQVprZnBFeFhsYUFsWitcclxuYT1zc3JjOjQ2ODM3NTU4IG1zaWQ6LSA0ZDc5YTM5MS01MzUxLTRmZDktYjdmOC0wODE2N2ViM2ExMTBcclxuYT1zc3JjOjQ2ODM3NTU4IG1zbGFiZWw6LVxyXG5hPXNzcmM6NDY4Mzc1NTggbGFiZWw6NGQ3OWEzOTEtNTM1MS00ZmQ5LWI3ZjgtMDgxNjdlYjNhMTEwXHJcbm09YXVkaW8gMzc1NzggVURQL1RMUy9SVFAvU0FWUEYgMTExIDEwMyA5IDAgOCAxMDUgMTMgMTEwIDExMyAxMjZcclxuYz1JTiBJUDQgMTkyLjE2OC4xLjEyXHJcbmE9cnRjcDo5IElOIElQNCAwLjAuMC4wXHJcbmE9Y2FuZGlkYXRlOjE0OTMzOTkxMzkgMSB1ZHAgMjExMzkzNzE1MSAxOTIuMTY4LjEuMTIgMzc1NzggdHlwIGhvc3QgZ2VuZXJhdGlvbiAwIG5ldHdvcmstY29zdCA5OTlcclxuYT1pY2UtdWZyYWc6OTZzQVxyXG5hPWljZS1wd2Q6N3ZxZDJlMnYwVXNGcW1ON3o3N2VEOUo1XHJcbmE9aWNlLW9wdGlvbnM6dHJpY2tsZVxyXG5hPWZpbmdlcnByaW50OnNoYS0yNTYgNjQ6REQ6OTg6RjU6MjU6MkU6MTQ6RUQ6MTk6MUE6RjY6ODg6NkQ6RTI6NEU6Qzk6OEU6OEQ6Mzc6Njg6ODg6NEM6QzE6RDg6QjE6NTk6REI6MzY6NUM6Njk6RTQ6RjdcclxuYT1zZXR1cDphY3RwYXNzXHJcbmE9bWlkOjFcclxuYT1leHRtYXA6MTQgdXJuOmlldGY6cGFyYW1zOnJ0cC1oZHJleHQ6c3NyYy1hdWRpby1sZXZlbFxyXG5hPWV4dG1hcDoyIGh0dHA6Ly93d3cud2VicnRjLm9yZy9leHBlcmltZW50cy9ydHAtaGRyZXh0L2Ficy1zZW5kLXRpbWVcclxuYT1leHRtYXA6NCBodHRwOi8vd3d3LmlldGYub3JnL2lkL2RyYWZ0LWhvbG1lci1ybWNhdC10cmFuc3BvcnQtd2lkZS1jYy1leHRlbnNpb25zLTAxXHJcbmE9ZXh0bWFwOjkgdXJuOmlldGY6cGFyYW1zOnJ0cC1oZHJleHQ6c2RlczptaWRcclxuYT1leHRtYXA6MTAgdXJuOmlldGY6cGFyYW1zOnJ0cC1oZHJleHQ6c2RlczpydHAtc3RyZWFtLWlkXHJcbmE9ZXh0bWFwOjExIHVybjppZXRmOnBhcmFtczpydHAtaGRyZXh0OnNkZXM6cmVwYWlyZWQtcnRwLXN0cmVhbS1pZFxyXG5hPXNlbmRyZWN2XHJcbmE9bXNpZDotIGM5NTZmMDY0LWM1NmQtNGU3ZS1hNGMyLTQ5ZmM2OGYyYTY4ZVxyXG5hPXJ0Y3AtbXV4XHJcbmE9cnRwbWFwOjExMSBvcHVzLzQ4MDAwLzJcclxuYT1ydGNwLWZiOjExMSB0cmFuc3BvcnQtY2NcclxuYT1mbXRwOjExMSBtaW5wdGltZT0xMDt1c2VpbmJhbmRmZWM9MVxyXG5hPXJ0cG1hcDoxMDMgSVNBQy8xNjAwMFxyXG5hPXJ0cG1hcDo5IEc3MjIvODAwMFxyXG5hPXJ0cG1hcDowIFBDTVUvODAwMFxyXG5hPXJ0cG1hcDo4IFBDTUEvODAwMFxyXG5hPXJ0cG1hcDoxMDUgQ04vMTYwMDBcclxuYT1ydHBtYXA6MTMgQ04vODAwMFxyXG5hPXJ0cG1hcDoxMTAgdGVsZXBob25lLWV2ZW50LzQ4MDAwXHJcbmE9cnRwbWFwOjExMyB0ZWxlcGhvbmUtZXZlbnQvMTYwMDBcclxuYT1ydHBtYXA6MTI2IHRlbGVwaG9uZS1ldmVudC84MDAwXHJcbmE9c3NyYzozNzM3NDg5OTg1IGNuYW1lOldDQVprZnBFeFhsYUFsWitcclxuYT1zc3JjOjM3Mzc0ODk5ODUgbXNpZDotIGM5NTZmMDY0LWM1NmQtNGU3ZS1hNGMyLTQ5ZmM2OGYyYTY4ZVxyXG5hPXNzcmM6MzczNzQ4OTk4NSBtc2xhYmVsOi1cclxuYT1zc3JjOjM3Mzc0ODk5ODUgbGFiZWw6Yzk1NmYwNjQtYzU2ZC00ZTdlLWE0YzItNDlmYzY4ZjJhNjhlXHJcbiJ9", &offer)

	// Set the remote SessionDescription
	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		panic(err)
	}

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// Output the answer in base64 so we can paste it in browser
	fmt.Println(utils.Encode(*peerConnection.LocalDescription()))

	// Block forever
	select {}
}
