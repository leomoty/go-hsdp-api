package iron_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTasksServices_GetTasks(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	tasks, resp, err := client.Tasks.GetTasks()
	if !assert.NotNil(t, resp) {
		return
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	if !assert.Nil(t, err) {
		return
	}
	if !assert.NotNil(t, tasks) {
		return
	}
	assert.Equal(t, 2, len(*tasks))
}

func TestTasksServices_GetTask(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	taskID := "bFp7OMpXdVsvRHp4sVtqb3gV"

	muxIRON.HandleFunc(apiProjectsPrefix+"/tasks/"+taskID, func(w http.ResponseWriter, r *http.Request) {
		if !assert.Equal(t, "GET", r.Method) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{
      "id": "`+taskID+`",
      "created_at": "2020-06-23T09:47:07.967Z",
      "updated_at": "2020-06-23T10:19:58.119Z",
      "project_id": "Bny3gFLzLlMrFFDrujopyocu",
      "code_id": "5e6640a5fbce220009c0385e",
      "code_history_id": "5e6640a5fbce220009c0385f",
      "status": "cancelled",
      "msg": "Cancelled via API.",
      "code_name": "loafoe/siderite",
      "code_rev": "1",
      "start_time": "2020-06-23T09:47:11.85Z",
      "end_time": "0001-01-01T00:00:00Z",
      "timeout": 3600,
      "payload": "mu4xSCwztB79NcmrJvFEdRnw0priIxMDxLPencrypted",
      "schedule_id": "5eebb5113de052000a93b1f5",
      "message_id": "6841477577898197071",
      "cluster": "9PbpheKmd0bSHIelR7O6ChcH"
    }`)
	})

	task, resp, err := client.Tasks.GetTask(taskID)
	if !assert.NotNil(t, resp) {
		return
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	if !assert.Nil(t, err) {
		return
	}
	if !assert.NotNil(t, task) {
		return
	}
	assert.Equal(t, taskID, task.ID)
}
