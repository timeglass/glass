package model

var DefaultConfig = &Config{
	CommitMessage: " [{{.}}]",
}

type Config struct {
	CommitMessage string `json:"commit_message"`
}
