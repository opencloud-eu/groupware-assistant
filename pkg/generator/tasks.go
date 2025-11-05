package generator

import (
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"strconv"

	"github.com/brianvoe/gofakeit/v7"
	"opencloud.eu/groupware-assistant/pkg/jmap"
	"opencloud.eu/groupware-assistant/pkg/tools"
)

func GenerateTasks(
	jmapUrl string,
	trace bool,
	color bool,
	username string,
	password string,
	accountId string,
	empty bool,
	tasklistId string,
	count uint,
	printer func(string),
) error {
	var s *jmap.TaskSender = nil
	{
		u, err := url.Parse(jmapUrl)
		if err != nil {
			return err
		}

		j, err := jmap.NewJmap(u, username, password, trace, color)
		if err != nil {
			return err
		}
		defer j.Close()

		s, err = jmap.NewTaskSender(j, accountId, tasklistId)
		if err != nil {
			return err
		}
	}
	defer s.Close()

	if empty {
		deleted, err := s.EmptyTasks()
		if err != nil {
			return err
		}
		if deleted > 0 {
			printer(fmt.Sprintf("🗑️ deleted %d tasks", deleted))
		} else {
			printer("ℹ️ did not delete any tasks, tasklist is empty")
		}
	}

	for i := range count {
		person := gofakeit.Person()
		task := map[string]any{
			"@type":       "Task",
			"version":     "1.0",
			"taskListIds": tools.ToBoolMap([]string{s.TaskList()}),
			"prodId":      tools.ProductName,
			"language":    tools.PickLanguage(),
		}

		task["@type"] = "individual"
		task["name"] = createName(person)
		nickNameId := id()
		task["nicknames"] = map[string]any{
			nickNameId: createNickName(person),
		}
		emails := map[string]any{}
		emailId := id()
		emails[emailId] = createEmail(person, 10)
		for i := range rand.Intn(3) {
			emails[id()] = createSecondaryEmail(gofakeit.Email(), i*100)
		}

		uid, err := s.CreateTask(task)
		if err != nil {
			return err
		}
		printer(fmt.Sprintf("📋 created %*s/%v uid=%v", int(math.Log10(float64(count))+1), strconv.Itoa(int(i+1)), count, uid))
	}
	return nil
}
