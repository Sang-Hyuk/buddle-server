package main

import (
	"github.com/sirupsen/logrus"
)

func main() {
	s, err := NewServer()
	if err != nil {
		logrus.Fatalf("Create server: %v", err)
	}

	if err := s.start(); err != nil {
		logrus.Fatalf("Start server: %v", err)
	}
}
