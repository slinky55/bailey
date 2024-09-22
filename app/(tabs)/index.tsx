import {ActivityIndicator, Alert, Button, StyleSheet, Text, TextInput, View} from 'react-native';
import {useEffect, useRef, useState} from "react";

import {mediaDevices, RTCIceCandidate, RTCPeerConnection, RTCView} from "react-native-webrtc";

// @ts-ignore
import DeviceNumber from 'react-native-device-number';

import rtcConfig from "@/config/rtc";

export default function HomeScreen() {
    const signalChannel = useRef<WebSocket | null>(null);
    const peerConnection = useRef<RTCPeerConnection | null>(null);
    const localMediaStream = useRef<MediaStream | null>(null);
    const remoteMediaStream = useRef<MediaStream | null>(null);
    const deviceNumber = useRef<string | null>(null);
    const iceBuffer = useRef<RTCIceCandidate[]>([]);

    const [calling, setCalling] = useState<string>('');
    const [onCall, setOnCall] = useState<boolean>(false);
    const [signaling, setSignaling] = useState<boolean>(false);

    async function _getDeviceNumber() {
        try {
            const number = await DeviceNumber.get();
            deviceNumber.current = number.mobileNumber;
        } catch (error) {
            console.log(error);
        }
    }

    function _initSignalConnection() {
        signalChannel.current = new WebSocket("wss://signal.crispycarnival.party/signal")
        signalChannel.current.onopen = () => {
            signalChannel.current?.send(deviceNumber.current!)
        }

        signalChannel.current.onmessage = async (e: MessageEvent) => {
            const incoming = JSON.parse(e.data);
            switch (incoming.type) {
                case "ice":
                    _handleIncomingIce(incoming);
                    break;
                case "offer":
                    await _handleIncomingOffer(incoming);
                    break;
                case "answer":
                    _handleIncomingAnswer(incoming);
                    break;
            }
        }
    }

    function _handleIncomingIce(incoming: any) {
        if (!peerConnection.current) {
            iceBuffer.current.push(new RTCIceCandidate(incoming.candidate));
        }

        peerConnection.current?.addIceCandidate(new RTCIceCandidate(incoming.candidate));
    }

    async function _handleIncomingOffer(incoming: any)  {
        setCalling(incoming.from);

        await _initMediaSources();
        await _initPc();

        const offerDescription = new RTCSessionDescription(incoming.offer);
        await peerConnection.current?.setRemoteDescription( offerDescription );

        const answerDescription = await peerConnection.current?.createAnswer();
        await peerConnection.current?.setLocalDescription(answerDescription);

        for (const candidate of iceBuffer.current) {
            peerConnection.current?.addIceCandidate(candidate);
        }
        iceBuffer.current = [];

        _sendAnswer(answerDescription);
    }

    async function _handleIncomingAnswer(incoming: any) {
        const answerDescription = new RTCSessionDescription(incoming.answer);
        await peerConnection.current?.setRemoteDescription(answerDescription);
    }

    function _sendIceCandidate(candidate: RTCIceCandidate) {
        signalChannel.current?.send(JSON.stringify({
            "type": "ice",
            "to": calling,
            "candidate": candidate,
        }));
    }

    function _sendOffer(offer: RTCSessionDescription) {
        signalChannel.current?.send(JSON.stringify({
            "type": "offer",
            "to": calling,
            "offer": offer,
        }));
    }

    function _sendAnswer(answer: RTCSessionDescription) {
        signalChannel.current?.send(JSON.stringify({
            "type": "answer",
            "to": calling,
            "answer": answer,
        }))
    }

    useEffect(() => {
        (async () => {
            await _getDeviceNumber();
            _initSignalConnection();
            setSignaling(true);
        })()

        return () => {
            signalChannel.current?.close();
            signalChannel.current = null;
            setSignaling(false);
        }
    }, [])

    async function _initMediaSources() {
        try {
            // @ts-ignore
            localMediaStream.current = await mediaDevices.getUserMedia(rtcConfig.mediaConstraints);
        } catch( err ) {
            Alert.alert('Error', 'Failed to get media sources', [
                {text: 'OK'},
            ]);
            console.log(err);
        }
    }

    async function _initPc() {
        peerConnection.current = new RTCPeerConnection(rtcConfig.peerConstraints);
        const pc = peerConnection.current;

        pc.addEventListener("connect", (e) => {
            switch (pc.connectionState) {
                case 'closed':
                    _closeCall();
                    break;
            }
        });

        pc.addEventListener("icecandidate", (e) => {
            if (!e.candidate) return;
            _sendIceCandidate(e.candidate)
        })

        pc.addEventListener("icecandidateerror", (e) => {
            console.error(e);
        })

        pc.addEventListener('iceconnectionstatechange', (e) => {
            switch(pc.iceConnectionState) {
                case 'connected':
                case 'completed':
                    setOnCall(true);
                    break;
            }
        });

        pc.addEventListener("negotiationneeded", async (e) => {
            const offerDescription = await pc.createOffer(rtcConfig.sessionConstraints);
            await pc.setLocalDescription(offerDescription);

            _sendOffer(offerDescription);
        })

        pc.addEventListener("track", (e) => {
            remoteMediaStream.current = remoteMediaStream.current || new MediaStream();
            // @ts-ignore
            remoteMediaStream.current.addTrack(e.track!);
        })

        localMediaStream.current?.getTracks().forEach(
            // @ts-ignore
            track => pc.addTrack(track, localMediaStream)
        );
    }

    async function _closeCall() {
        peerConnection.current?.close();
        peerConnection.current = null;
        setOnCall(false);
    }

    if (!signaling) {
        return (
            <View style={styles.container}>
                <ActivityIndicator></ActivityIndicator>
            </View>
        );
    }

    return (
        <View style={styles.container}>
            {!onCall && (
                <>
                    <Text style={styles.title}>Dial Pad</Text>
                    <TextInput
                        style={styles.input}
                        placeholder="Enter number"
                        keyboardType="phone-pad"
                        value={calling}
                        onChangeText={setCalling}
                    />
                    <Button title="Dial" onPress={_getDeviceNumber} />
                </>
            )}
            {onCall && (
                <>
                    <Text style={styles.title}>On call with {calling}</Text>
                    <Button title="End Call" onPress={() => {}} />
                    <RTCView streamURL={remoteMediaStream.current?.id}/>
                </>
            )}
        </View>
    )
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        padding: 20,
        backgroundColor: '#f5f5f5',
    },
    title: {
        fontSize: 24,
        marginBottom: 20,
    },
    input: {
        height: 50,
        width: '100%',
        borderColor: '#ccc',
        borderWidth: 1,
        paddingHorizontal: 10,
        marginBottom: 20,
    },
    status: {
        fontSize: 18,
        marginBottom: 20,
    },
});
