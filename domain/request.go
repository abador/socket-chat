package domain

type Request struct {
	Action int  `json:"action"`
	Data   Data `json:"data"`
}
