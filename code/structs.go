package main

type DtInfoEvent struct {
	EventType string `json:"eventType"`
	Source    string `json:"source"`
	//Start            int               `json: "start,omitempty"`
	//End              int               `json: "end,omitempty"`
	AttachRules      DtAttachRules     `json:"attachRules"`
	CustomProperties map[string]string `json:"customProperties"`
	Description      string            `json:"description"`
	Title            string            `json:"title"`
}

type DtAttachRules struct {
	TagRule []DtTagRule `json:"tagRule"`
}

// DtTag defines a Dynatrace configuration structure
type DtTag struct {
	Context string `json:"context"`
	Key     string `json:"key"`
	Value   string `json:"value,omitempty"`
}

// DtTagRule defines a Dynatrace configuration structure
type DtTagRule struct {
	MeTypes []string `json:"meTypes"`
	Tags    []DtTag  `json:"tags"`
}

// Zendesk structs
type ZDTicket struct {
	Request ZDRequest `json:"request"`
}

type ZDRequest struct {
	Requester ZDRequester `json:"requester"`
	Subject   string      `json:"subject"`
	Comment   ZDComment   `json:"comment"`
	Tags      []string    `json:"tags"`
}

type ZDRequester struct {
	Name string `json:"name"`
}

type ZDComment struct {
	HTMLBody string `json:"html_body"`
}

// Model the API response object
type ZDTicketResponse struct {
	Request ZDResponseRequest `json:"request"`
}

type ZDResponseRequest struct {
	Description string `json:"description"`
	ID          int    `json:"id"`
	Status      string `json:"status"`
	Subject     string `json:"subject"`
}
