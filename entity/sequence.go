package entity

type Sequence struct {
    Steps       []Step `json:"steps"`
    Subscribers int    `json:"subscribers"`
}
