package errors

import "fmt"

var _ error = &TaskListNotOwnnedByHostError{}

type TaskListNotOwnnedByHostError struct {
	OwnedByIdentity string
	MyIdentity      string
	TasklistName    string
}

func (m *TaskListNotOwnnedByHostError) Error() string {
	return fmt.Sprintf("task list is not owned by this host: OwnedBy: %s, Me: %s, Tasklist: %s",
		m.OwnedByIdentity, m.MyIdentity, m.TasklistName)
}

func NewTaskListNotOwnnedByHostError(ownedByIdentity string, myIdentity string, tasklistName string) *TaskListNotOwnnedByHostError {
	return &TaskListNotOwnnedByHostError{
		OwnedByIdentity: ownedByIdentity,
		MyIdentity:      myIdentity,
		TasklistName:    tasklistName,
	}
}
