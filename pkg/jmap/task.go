package jmap

import (
	"fmt"
	"slices"
	"strings"

	"opencloud.eu/groupware-assistant/pkg/tools"
)

type TaskRights struct {
	MayReadItems     bool `json:"mayReadItems"`
	MayWriteAll      bool `json:"mayWriteAll"`
	MayWriteOwn      bool `json:"mayWriteOwn"`
	MayUpdatePrivate bool `json:"mayUpdatePrivate"`
	MayRSVP          bool `json:"mayRSVP"`
	MayAdmin         bool `json:"mayAdmin"`
	MayDelete        bool `json:"mayDelete"`
}

type TaskList struct {
	Id               string            `json:"id,omitempty"`
	Role             string            `json:"role,omitempty"`
	Name             string            `json:"name,omitempty"`
	Description      string            `json:"description,omitempty"`
	Color            string            `json:"color,omitempty"`
	KeywordColors    map[string]string `json:"keywordColors,omitempty"`
	CategoryColors   map[string]string `json:"categoryColors,omitempty"`
	SortOrder        int               `json:"sortOrder,omitzero"`
	IsSubscribed     bool              `json:"isSubscribed,omitzero"`
	TimeZone         string            `json:"timeZone,omitempty"`
	WorkflowStatuses []string          `json:"workflowStatuses,omitempty"`
	MyRights         *TaskRights       `json:"myRights,omitempty"`
	//ShareWith map[string]TaskRights `json:"shareWith,omitempty"`
}

type NewTaskList struct {
	Role             string            `json:"role,omitempty"`
	Name             string            `json:"name,omitempty"`
	Description      string            `json:"description,omitempty"`
	Color            string            `json:"color,omitempty"`
	KeywordColors    map[string]string `json:"keywordColors,omitempty"`
	CategoryColors   map[string]string `json:"categoryColors,omitempty"`
	SortOrder        int               `json:"sortOrder,omitzero"`
	IsSubscribed     bool              `json:"isSubscribed,omitzero"`
	TimeZone         string            `json:"timeZone,omitempty"`
	WorkflowStatuses []string          `json:"workflowStatuses,omitempty"`
}

func ListTasklists(j *Jmap, accountId string) ([]TaskList, error) {
	if accountId, err := j.taskAccountId(accountId); err != nil {
		return nil, err
	} else {
		if list, err := objects[TaskList](j, accountId, TaskListsObjectType, JmapTasks); err != nil {
			return nil, err
		} else {
			slices.SortFunc(list, func(a, b TaskList) int { return strings.Compare(a.Id, b.Id) })
			return list, err
		}
	}
}

func CreateTasklist(j *Jmap, accountId string, tasklist NewTaskList) (string, error) {
	if accountId, err := j.taskAccountId(accountId); err != nil {
		return "", err
	} else {
		if m, err := tools.Remap(tasklist); err != nil {
			return "", err
		} else {
			return create(j, TaskListsObjectType, createBody(accountId, TaskListsObjectType, JmapTasks, m))
		}
	}
}

func DeleteTasklist(j *Jmap, accountId string, id string) error {
	if accountId, err := j.taskAccountId(accountId); err != nil {
		return err
	} else {
		return destroy(j, accountId, TaskListsObjectType, JmapTasks, []string{id})
	}
}

type TaskSender struct {
	j          *Jmap
	accountId  string
	tasklistId string
}

func (s *TaskSender) TaskList() string {
	return s.tasklistId
}

func NewTaskSender(j *Jmap, accountId string, tasklistId string) (*TaskSender, error) {
	if accountId, err := j.taskAccountId(accountId); err != nil {
		return nil, err
	} else {
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
			return nil, fmt.Errorf("failed to find a %s with role=inbox", TaskListsObjectType)
		}

		return &TaskSender{
			j:          j,
			accountId:  accountId,
			tasklistId: tasklistId,
		}, nil
	}
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
	return create(s.j, TaskObjectType, createBody(s.accountId, TaskObjectType, JmapTasks, c))
}
