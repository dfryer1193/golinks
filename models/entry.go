package models

type Entry struct {
	Path   string `json:"path"`
	Target string `json:"target"`
}

type UpdateDelta struct {
	Old *Entry `json:"old"`
	New *Entry `json:"new"`
}
