package serialutil

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	serialLength = 6
)

func GenerateId(prefix string) (string, error) {
	id, err := gonanoid.Generate("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", serialLength)
	if err != nil {
		return "", err
	}
	return prefix + "-" + id, nil
}
