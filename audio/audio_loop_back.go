package audio

import (
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"time"

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
			loopBack.Capture.Channels = 2
			loopBack.Playback.Channels = 2
			playback.Playback.Channels = 2
			playback.Playback.DeviceID = info.ID.Pointer()
		}
	}

	// var playbackSampleCount uint32
	// var capturedSampleCount uint32
	// pCapturedSamples := make([]byte, 0)

	// sizeInBytes := uint32(malgo.SampleSizeInBytes(loopBack.Capture.Format))
	const sampleRate = 48000
	const channels = 2 // mono; 2 for stereo

	enc, err := opus.NewEncoder(sampleRate, channels, opus.AppAudio)
	if err != nil {
		panic(fmt.Sprintf("can not create encoder: %s", err))
	}

	// encodedDataTotal := make([]int16, 0)
	// consumed := 0

	// dec, err := opus.NewDecoder(sampleRate, channels)
	// if err != nil {
	// }
	// pcm := make([]int16, int(480*2))

	prev := time.Now()
	// make(chan []byte)
	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		start := time.Now()

		fmt.Printf("lat: %s\n", time.Since(prev))
		prev = start
		// sampleCount := framecount * loopBack.Capture.Channels * sizeInBytes

		// newCapturedSampleCount := capturedSampleCount + sampleCount

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

		// start := time.Now()
		n, err := enc.Encode(dataToEncode, encodedData)
		// fmt.Printf("e: %d\n", (int64(time.Now())-int64(start)) /time.Millisecond)

		if err != nil {
			fmt.Println(fmt.Sprintf("can not encode %s", err))
			return
		}
		encodedData = encodedData[:n]

		// _, errr := dec.Decode(encodedData, pcm)
		// if errr != nil {
		// 	fmt.Println(fmt.Sprintf("can not decode %s", errr))
		// }
		// encodedDataTotal = append(encodedDataTotal, pcm...)

		// consumed++
		// fmt.Printf("consumed: %d", consumed)
		// if consumed == 1000 {
		// 	b := make([]byte, 0)
		// 	for _, curD := range encodedDataTotal {
		// 		bb := make([]byte, 2)
		// 		binary.LittleEndian.PutUint16(bb, uint16(curD))
		// 		b = append(b, bb...)
		// 	}
		// 	ioutil.WriteFile("out2.pcm", b, 0644)
		// }

		// fmt.Printf("%f\n", frameSizeMs)
		// pCapturedSamples = append(pCapturedSamples, pSample...)
		samples <- SampleData{N: int32(frameSizeMs), Samples: encodedData}
		// capturedSampleCount = newCapturedSampleCount

	}

	// go func() {

	// }

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

	// fmt.Println("Press Enter to stop recording...")
	// fmt.Scanln()

	// device.Uninit()

	// onSendFrames := func(pSample, nil []byte, framecount uint32) {
	// 	samplesToRead := framecount * playback.Playback.Channels * sizeInBytes
	// 	if samplesToRead > capturedSampleCount-playbackSampleCount {
	// 		samplesToRead = capturedSampleCount - playbackSampleCount
	// 	}

	// 	copy(pSample, pCapturedSamples[playbackSampleCount:playbackSampleCount+samplesToRead])

	// 	playbackSampleCount += samplesToRead
	// 	if playbackSampleCount == uint32(len(pCapturedSamples)) {
	// 		playbackSampleCount = 0
	// 	}
	// }

	// fmt.Println("Playing...")
	// playbackCallbacks := malgo.DeviceCallbacks{
	// 	Data: onSendFrames,
	// }

	// device, err = malgo.InitDevice(ctx.Context, playback, playbackCallbacks)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	// err = device.Start()
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	return samples
}
