package cli

import (
	"testing"

	"github.com/urfave/cli"
)

// For testing filter out admin operations
type AdminTestAuthorizationManager struct {
}

//  For testing filter out admin operations
func NewAdminTestAuthorizationManager() *AdminTestAuthorizationManager {
	return &AdminTestAuthorizationManager{}
}

//Decides which operation will be available in this CLI
func (m *AdminTestAuthorizationManager) FilterUnauthorizedOperations(app *cli.App) *cli.App {
	newCmds := []cli.Command{}
	for _, cmd := range app.Commands {
		if cmd.Name != "admin" {
			newCmds = append(newCmds, cmd)
		}
	}
	app.Commands = newCmds
	return app
}

func TestAdminTestAuthorizationManager(t *testing.T) {
	SetAuthorizationManager(NewAdminTestAuthorizationManager())
	app := NewCliApp()
	for _, cmd := range app.Commands {
		if cmd.Name == "admin" {
			t.Fail()
		}
	}
}
