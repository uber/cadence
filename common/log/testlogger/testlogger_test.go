package testlogger

import (
	"fmt"
	"testing"
	"time"
)

var done = make(chan struct{})

func TestMain(m *testing.M) {
	m.Run()
	close(done)
	time.Sleep(100 * time.Millisecond)
}
func TestABefore(t *testing.T) {
	go func() {
		t.Logf("sleeping")
		time.Sleep(10 * time.Millisecond)
		<-done
		t.Logf("too late")
		fmt.Println("goroutine done") // prove it ran
	}()
	time.Sleep(time.Millisecond)
}

func TestZLater(t *testing.T) {
	t.Logf("another test")
	time.Sleep(time.Second)
}
