package main

// Handshake type is message for communicate with aggregate server
type Handshake struct {
	Plan string `json:"action"`
	Channel string `json:"user"`
}

// Message type is message from receiving approval result from aggregate server
type Message struct {
	Action string `json:"action"`
	User   string `json:"user"`
}

