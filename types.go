package main

type Handshake struct {
	Plan string `json:"action"`
	Channel string `json:"user"`
}

type Message struct {
	Action string `json:"action"`
	User   string `json:"user"`
}

