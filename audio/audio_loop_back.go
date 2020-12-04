package audio

import (
	"fmt"
	"os"
	"regexp"
	"unsafe"

	"github.com/gen2brain/malgo"
	"github.com/hraban/opus"
)

// StartCapture capturing audio from default playback device
func StartCapture() {
	ctx, err := malgo.InitContext([]malgo.Backend{malgo.BackendWasapi}, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	infos, err := ctx.Devices(malgo.Playback)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	loopBack := malgo.DefaultDeviceConfig(malgo.Loopback)
	loopBack.Capture.Format = malgo.FormatS16
	loopBack.Playback.Format = malgo.FormatS16
	loopBack.SampleRate = 48000

	playback := malgo.DefaultDeviceConfig(malgo.Playback)
	playback.Playback.Format = malgo.FormatS16

	playback.SampleRate = 48000

	fmt.Println("Playback Devices")
	re, err := regexp.Compile(`USB`)

	for i, info := range infos {
		e := "ok"
		full, err := ctx.DeviceInfo(malgo.Playback, info.ID, malgo.Shared)
		if err != nil {
			e = err.Error()
		}
		if re.FindString(info.Name()) != "" {
			fmt.Printf("    %d: %v, %s, [%s], channels: %d-%d, samplerate: %d-%d\n",
				i, info.ID, info.Name(), e, full.MinChannels, full.MaxChannels, full.MinSampleRate, full.MaxSampleRate)
			loopBack.Capture.DeviceID = info.ID.Pointer()
			loopBack.Capture.Channels = full.MaxChannels
			loopBack.Playback.Channels = full.MaxChannels
			playback.Playback.Channels = full.MaxChannels
			playback.Playback.DeviceID = info.ID.Pointer()
		}
	}

	var playbackSampleCount uint32
	var capturedSampleCount uint32
	pCapturedSamples := make([]byte, 0)

	sizeInBytes := uint32(malgo.SampleSizeInBytes(loopBack.Capture.Format))
	const sampleRate = 48000
	const channels = 2 // mono; 2 for stereo

	enc, err := opus.NewEncoder(sampleRate, channels, opus.AppAudio)
	if err != nil {
		panic(fmt.Sprintf("can not create encoder: %s", err))
	}

	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		sampleCount := framecount * loopBack.Capture.Channels * sizeInBytes

		newCapturedSampleCount := capturedSampleCount + sampleCount

		frameSizeMs := float32(framecount) * 1000 / sampleRate
		switch frameSizeMs {
		case 2.5, 5, 10, 20, 40, 60:
			// Good.
		default:
			panic(fmt.Sprintf("Illegal frame size: (%f ms)", frameSizeMs))
		}

		encodedData := make([]byte, 1000)
		n, err := enc.Encode(*(*[]int16)(unsafe.Pointer(&pSample)), encodedData)
		if err != nil {
			panic(fmt.Sprintf("can not encode %s", err))
		}
		encodedData = encodedData[:n]
		fmt.Printf("%d\n", n)
		pCapturedSamples = append(pCapturedSamples, pSample...)

		capturedSampleCount = newCapturedSampleCount

	}

	fmt.Println("Recording...")
	captureCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}
	device, err := malgo.InitDevice(ctx.Context, loopBack, captureCallbacks)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = device.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Press Enter to stop recording...")
	fmt.Scanln()

	device.Uninit()

	onSendFrames := func(pSample, nil []byte, framecount uint32) {
		samplesToRead := framecount * playback.Playback.Channels * sizeInBytes
		if samplesToRead > capturedSampleCount-playbackSampleCount {
			samplesToRead = capturedSampleCount - playbackSampleCount
		}

		copy(pSample, pCapturedSamples[playbackSampleCount:playbackSampleCount+samplesToRead])

		playbackSampleCount += samplesToRead
		if playbackSampleCount == uint32(len(pCapturedSamples)) {
			playbackSampleCount = 0
		}
	}

	fmt.Println("Playing...")
	playbackCallbacks := malgo.DeviceCallbacks{
		Data: onSendFrames,
	}

	device, err = malgo.InitDevice(ctx.Context, playback, playbackCallbacks)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = device.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Press Enter to quit...")
	fmt.Scanln()

	device.Uninit()
}
