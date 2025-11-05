package jmap

import (
	"fmt"
)

var TaskListsObjectType = "TaskLists"
var TaskObjectType = "Task"

type TaskSender struct {
	j          *Jmap
	accountId  string
	tasklistId string
}

func (s *TaskSender) TaskList() string {
	return s.tasklistId
}

func NewTaskSender(j *Jmap, accountId string, tasklistId string) (*TaskSender, error) {
	if accountId == "" {
		// use default mail account
		accountId = j.session.PrimaryAccounts.Tasks
		if accountId == "" {
			return nil, fmt.Errorf("session has no matching primary account")
		}
	} else {
		if _, ok := j.session.Accounts[accountId]; !ok {
			return nil, fmt.Errorf("account ID '%s' does not exist in session", accountId)
		}
	}

	tasklistsById, err := objectsById(j, accountId, TaskListsObjectType, JmapTasks)
	if err != nil {
		return nil, err
	}
	if tasklistId != "" {
		if _, ok := tasklistsById[tasklistId]; !ok {
			return nil, fmt.Errorf("tasklist with id '%s' does not exist", tasklistId)
		}
	} else {
		for id, tasklist := range tasklistsById {
			if role, ok := tasklist["role"]; ok {
				if role.(string) == "inbox" {
					tasklistId = id
					break
				}
			}
		}
	}
	if tasklistId == "" {
		return nil, fmt.Errorf("failed to find an inbox TaskList")
	}

	return &TaskSender{
		j:          j,
		accountId:  accountId,
		tasklistId: tasklistId,
	}, nil
}

func (s *TaskSender) Close() error {
	return nil
}

func (s *TaskSender) EmptyTasks() (uint, error) {
	return empty(s.j, s.accountId, TaskObjectType, JmapTasks, map[string]any{
		"inTaskList": s.tasklistId,
	}, s.destroy)
}

func (s *TaskSender) destroy(ids []string) error {
	return destroy(s.j, s.accountId, TaskObjectType, JmapTasks, ids)
}

func (s *TaskSender) CreateTask(c map[string]any) (string, error) {
	body := map[string]any{
		"using": []string{JmapCore, JmapTasks},
		"methodCalls": []any{
			[]any{
				TaskObjectType + "/set",
				map[string]any{
					"accountId": s.accountId,
					"create": map[string]any{
						"c": c,
					},
				},
				"0",
			},
		},
	}
	return create(s.j, "c", TaskObjectType, body)
}
