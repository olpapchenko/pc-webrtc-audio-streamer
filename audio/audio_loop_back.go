package audio

import (
	"encoding/binary"
	"fmt"
	"os"
	"regexp"

	"github.com/gen2brain/malgo"
	"github.com/hraban/opus"
)

// SampleData contains sample encoded data
type SampleData struct {
	N       int32
	Samples []byte
}

// StartCapture capturing audio from default playback device
func StartCapture() <-chan SampleData {
	ctx, err := malgo.InitContext([]malgo.Backend{malgo.BackendWasapi}, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})

	samples := make(chan SampleData, 5000)

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

	const sampleRate = 48000
	const channels = 2

	loopBack := malgo.DefaultDeviceConfig(malgo.Loopback)
	loopBack.Capture.Format = malgo.FormatS16
	loopBack.Playback.Format = malgo.FormatS16
	loopBack.SampleRate = sampleRate

	playback := malgo.DefaultDeviceConfig(malgo.Playback)
	playback.Playback.Format = malgo.FormatS16

	playback.SampleRate = sampleRate

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
			loopBack.Capture.Channels = channels
			loopBack.Playback.Channels = channels
			playback.Playback.Channels = channels
			playback.Playback.DeviceID = info.ID.Pointer()
		}
	}

	enc, err := opus.NewEncoder(sampleRate, channels, opus.AppAudio)
	if err != nil {
		panic(fmt.Sprintf("can not create encoder: %s", err))
	}

	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {

		frameSizeMs := float32(framecount) * 1000 / sampleRate
		switch frameSizeMs {
		case 2.5, 5, 10, 20, 40, 60:
			// Good.
		default:
			fmt.Printf("bad frame size: %f", frameSizeMs)
			return
		}

		encodedData := make([]byte, 5000)
		dataToEncode := make([]int16, 0)

		for i := 0; i < len(pSample); i += 2 {
			dataToEncode = append(dataToEncode, int16(binary.LittleEndian.Uint16(pSample[i:i+2])))
		}

		n, err := enc.Encode(dataToEncode, encodedData)

		if err != nil {
			fmt.Println(fmt.Sprintf("can not encode %s", err))
			return
		}
		encodedData = encodedData[:n]

		samples <- SampleData{N: int32(frameSizeMs), Samples: encodedData}
	}

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

	return samples
}
