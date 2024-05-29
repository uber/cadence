package testlogger

import (
	"testing"
	"time"
)

var done = make(chan struct{})

func TestMain(m *testing.M) {
	m.Run()
	time.Sleep(time.Second)
}

func TestALogFlaky(t *testing.T) {
	go func() {
		t.Logf("sleeping")
		time.Sleep(10 * time.Millisecond)
		t.Logf("too late")
	}()
	time.Sleep(time.Millisecond)
}

func TestZLater(t *testing.T) {
	t.Logf("another test")
}
