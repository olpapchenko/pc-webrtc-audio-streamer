# pc-webrtc-audio-streamer
The project allows streaming the sound from Windows PC to remote devices via WebRTC.

The project consists of a server written on go. The server can be started on the host Windows PC, capture all sound, and transmit via WebRTC channel to another peer.
This can be useful in cases, when you want to plug your Bluetooth headphones, but have no Bluetooth adapter installed. In this case, you can  stream PC audio to the
mobile phone and plug in your headphones to it.

### How to connect.
1. start the server locally.
2. on mobile phone in Chrome mobile browser open http://local_ip_of_your_windows:3000/static
3. click the "Start" button to start the audio stream.
**Note:** your mobile phone should be connected to the same local network as your PC.

![UseCase](UseCase.jpg?raw=true)

### How to build.
```sh
 go build main.go
```

### Pre-built bin files
For pre-built exe file refer to releases section.

**Note:** bin file is so big - because go runtime is packed in it.

### Used libs:
* [malgo](https://github.com/gen2brain/malgo)
* [pion](https://github.com/pion/webrtc)
