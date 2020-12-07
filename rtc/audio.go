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
func StartWebRtc(sampleData <-chan audio.SampleData, session string) string {

	// Create a new RTCPeerConnection
	m := webrtc.MediaEngine{}

	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 16000, Channels: 2, SDPFmtpLine: "stereo=1", RTCPFeedback: nil},
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
			sampleDuration := time.Millisecond * time.Duration(int64(data.N))

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
	utils.Decode(session, &offer)

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
	return utils.Encode(*peerConnection.LocalDescription())
}
