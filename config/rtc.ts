const rtcConfig = {
    mediaConstraints: {
        audio: true,
        video: false,
    },
    peerConstraints: {
        iceServers: [
            {
                urls: 'stun:stun.l.google.com:19302'
            }
        ]
    },
    sessionConstraints: {
        mandatory: {
            OfferToReceiveAudio: true,
            OfferToReceiveVideo: false,
            VoiceActivityDetection: true
        }
    }
}

export default rtcConfig;