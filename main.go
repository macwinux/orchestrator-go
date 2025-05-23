package main

import (
	"fmt"
	"log"
	//"os"
	"scheduler/manager"
	"scheduler/task"
	"scheduler/worker"
	//"strconv"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {
	//host := os.Getenv("CUBE_HOST")
	//port, _ := strconv.Atoi(os.Getenv("CUBE_PORT"))
	host := "localhost"
	port := 5555
	fmt.Println("Starting Cube Worker")

	// Instanciamos el worker
	w := worker.Worker{
		Queue: *queue.New(),
		Db: make(map[uuid.UUID]*task.Task),
	}
	api := worker.Api{Address: host, Port: port, Worker: &w}
	go runTasks(&w)
	// Use in a docker image
	//go w.CollectStats()
	go api.Start()

	//creamos el manager pasandole el worker en formato slice
	worker := []string{fmt.Sprintf("%s:%d", host, port)}
	m := manager.New(worker)

	for i := range 3 {
		t := task.Task{
			ID: uuid.New(),
			Name: fmt.Sprintf("test-container-%d", i),
			State: task.Scheduled,
			Image: "strm/helloworld-http",
		}
		te := task.TaskEvent {
			ID: uuid.New(),
			State: task.Running,
			Task: t,
		}
		m.AddTask(te)
		m.SendWork()
	}

	go func() {
		for {
			fmt.Printf("[Manager] Updating tasks from %d workers\n", len(m.Workers))
			m.UpdateTasks()
			time.Sleep(15*time.Second)
		}
	}()

	for {
		for _, t := range m.TaskDb {
			fmt.Printf("[Manager] Task: id: %s, state: %d\n", t.ID, t.State)
			time.Sleep(15 * time.Second)
		}
	}

}

func runTasks(w *worker.Worker) {
	for {
		if w.Queue.Len() != 0 {
			result := w.RunTask()
			if result.Error != nil {
				log.Printf("Error running task: %v\n", result.Error)
			}
		} else {
			log.Printf("No tasks to process currently.\n")
		}
		log.Println("Sleeping for 10 seconds.")
		time.Sleep(10 * time.Second)
	}
}