package domain

type Response struct {
	Type      int            `json:"type"`
	Objects   []ResponseData `json:"objects"`
	Message   string
	OrigError string
}

type ResponseData struct {
	Type   int         `json:"type"`
	Object interface{} `json:"object"`
}

type ResponseGeneric struct {
	Type      int `json:"type"`
	Message   string
	OrigError string
}
