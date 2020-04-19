package config

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

func UniqueID() string {
	hn, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	h := md5.New()
	io.WriteString(h, hn)
	return fmt.Sprintf("%x", h.Sum(nil))
}
