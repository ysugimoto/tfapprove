package main

// Handshake type is message for communicate with aggregate server
type Handshake struct {
	Plan    string `json:"plan"`
	Channel string `json:"channel"`
}

// Message type is message from receiving approval result from aggregate server
type Message struct {
	Action string `json:"action"`
	User   string `json:"user"`
}

type Action struct {
	Type string `json:"type"`
}
