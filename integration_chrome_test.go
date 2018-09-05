// +build integration

package webrtc

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/pions/webrtc/pkg/datachannel"
	"github.com/pions/webrtc/pkg/ice"

	"github.com/pions/webrtc/internal/integration/chrome"
	"github.com/pions/webrtc/internal/integration/server"
	"github.com/stretchr/testify/assert"
)

func TestDataChannel_Create(t *testing.T) {
	s := server.New()
	defer s.Close()

	err := s.Spawn()
	if err != nil {
		panic(err)
	}

	c := chrome.New()
	defer c.Close()

	err = c.Spawn()
	if err != nil {
		panic(err)
	}

	err = c.Page("static/datachannel_create.html")
	if err != nil {
		panic(err)
	}

	config := RTCConfiguration{}

	// Create a new RTCPeerConnection
	peerConnection, err := New(config)
	check(err)

	// Create a datachannel with label 'data'
	dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	check(err)

	peerConnection.OnICEConnectionStateChange = func(connectionState ice.ConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())

		// TODO: find the correct place for this
		if connectionState == ice.ConnectionStateConnected {
			fmt.Println("sending openchannel")
			err := dataChannel.SendOpenChannelMessage()
			if err != nil {
				fmt.Println("faild to send openchannel", err)
			}
		}
	}

	// Create an offer to send to the browser
	offer, err := peerConnection.CreateOffer(nil)
	check(err)

	offerData, err := json.Marshal(offer)
	check(err)

	s.Signal(offerData)

	answerData := s.OnSignal()

	var answer RTCSessionDescription
	err = json.Unmarshal(answerData, &answer)
	check(err)

	// Apply the answer as the remote description
	err = peerConnection.SetRemoteDescription(answer)
	check(err)

	go func() {
		for {
			time.Sleep(5 * time.Second)
			err := dataChannel.Send(datachannel.PayloadString{Data: []byte("message")})
			check(err)
		}
	}()

	reply := s.Listen()

	assert.Equal(t, reply.Context, "onmessage")
	assert.Equal(t, reply.Message, "message")
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
