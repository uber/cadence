package testlogger

import (
	"fmt"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	m.Run()
	time.Sleep(100 * time.Millisecond)
}
func TestABefore(t *testing.T) {
	go func() {
		t.Logf("sleeping")
		time.Sleep(10 * time.Millisecond)
		t.Logf("too late")
		fmt.Println("goroutine done") // prove it ran
	}()
	time.Sleep(time.Millisecond)
}

func TestZLater(t *testing.T) {
	t.Logf("another test")
	time.Sleep(time.Second)
}
