package errors

import "fmt"

type ShutdownError struct {
	Service string
	Message string
}

func (m *ShutdownError) Error() string {
	return fmt.Sprintf("shutdown error: service %s, %s", m.Service, m.Message)
}
