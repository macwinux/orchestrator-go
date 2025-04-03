package main

import (
	"context"
	"fmt"
	"scheduler/manager"
	"scheduler/node"
	"scheduler/task"
	"scheduler/worker"
	"time"

	"github.com/docker/docker/client"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func createContainer(ctx context.Context) (*task.Docker, *task.DockerResult) {
	c := task.Config{
		Name: "test-container-1",
		Image: "postgres:13",
		Env: []string{
			"POSTGRES_USER=cube",
			"POSTGRES_PASSWORD=secret",
		},
	}
	dc, _ :=client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	d := task.Docker{
		Client: dc,
		Config: c,
	}

	result := d.Run(ctx)
	if result.Error != nil {
		fmt.Printf("%v\n", result.Error)
		return nil, nil
	}
	fmt.Printf("Container %s is running with config %v\n", result.ContainerId, c)
	return &d, &result
}

func stopContainer(d *task.Docker, id string, ctx context.Context) *task.DockerResult {
	result := d.Stop(id, ctx)
	if result.Error != nil {
		fmt.Printf("%v\n", result.Error)
		return nil
	}

	fmt.Printf("Container %s has been stopped and removed\n", result.ContainerId)
	return &result
}

func main() {
	t := task.Task{
		ID: uuid.New(),
		Name: "Task-1",
		State: task.Pending,
		Image: "Image-1",
		Memory: 1024,
		Disk: 1,
	}

	te := task.TaskEvent{
		ID: uuid.New(),
		State: task.Pending,
		Timestamp: time.Now(),
		Task: t,
	}

	fmt.Printf("task: %v\n", t)
	fmt.Printf("taskEvent: %v\n", te)

	w := worker.Worker{
		Name: "worker-1",
		Queue: *queue.New(),
		Db: make(map[uuid.UUID]*task.Task),
	}
	fmt.Printf("worker: %v\n", w)
	w.CollectStats()
	w.RunTask()
	w.StartTask()
	w.StopTask()

	m := manager.Manager{
		Pending: *queue.New(),
		TaskDb: make(map[string][]*task.Task),
		EventDb: make(map[string][]*task.TaskEvent),
		Workers: []string{w.Name},
	}
	fmt.Printf("manager: %v\n", m)
	m.SelectWorker()
	m.UpdateTasks()
	m.SendWork()

	n := node.Node{
		Name: "Node-1",
		Ip: "192.168.1.1",
		Cores: 4,
		Memory: 1024,
		Disk: 25,
		Role: "worker",
	}

	fmt.Printf("node: %v\n", n)

	fmt.Printf("create a test container\n")
	ctx := context.Background()
	dockerTask, createResult := createContainer(ctx)

	time.Sleep(time.Second * 5)

	fmt.Printf("stopping container %s\n", createResult.ContainerId)
	_ = stopContainer(dockerTask, createResult.ContainerId, ctx)
}