/* eslint-env browser */

let pc = new RTCPeerConnection({                         
    iceServers: [
      {     
          urls: []
      }                    
    ]                                             
  }) 
  let log = msg => {
    document.getElementById('div').innerHTML += msg + '<br>'
  }
  
  pc.ontrack = function (event) {
    var el = document.createElement(event.track.kind)
    el.srcObject = event.streams[0]
    el.autoplay = true
    el.controls = true
  
    document.getElementById('remoteAudio').appendChild(el)
  }
  
  pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
  pc.onicecandidate = event => {
    if (event.candidate === null) {
        console.log("fetching session...");
        console.log("local: " + btoa(JSON.stringify(pc.localDescription)));
        fetch('/session/', {
          method: 'POST',
          body: JSON.stringify({
            session: btoa(JSON.stringify(pc.localDescription))
          })
        })
        .then(response => {
          console.log("Session fetched");
          response.text().then(desc => {
            setRemoteDesciprtion(desc);
          });
        }).catch(err => {
          console.log(err);
          alert("Error occurred: " + error);
        });

       
    }
  }
  
  pc.addTransceiver('audio', {'direction': 'sendrecv'})
  
  pc.createOffer().then(d => pc.setLocalDescription(d)).catch(log)
  
  function setRemoteDesciprtion(sd) {
    if (sd === '') {
      return alert('Session Description must not be empty')
    }
  
    try {
      pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(atob(sd))))
    } catch (e) {
      console.log("error occured while setting remode desc: " + e);
      alert(e)
    }
  }
  