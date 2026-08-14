package main

import "testing"

func TestBanner(t *testing.T) {
	if banner() != "healer alive" {
		t.Fatal("unexpected banner")
	}
}
