package main

import (
	"fmt"
	"time"
)

// Приложение эмулирует получение и обработку неких тасков. Пытается и получать, и обрабатывать в многопоточном режиме.
// Приложение должно генерировать таски 10 сек. Каждые 3 секунды должно выводить в консоль результат всех обработанных к этому моменту тасков (отдельно успешные и отдельно с ошибками).

// ЗАДАНИЕ: сделать из плохого кода хороший и рабочий - as best as you can.
// Важно сохранить логику появления ошибочных тасков.
// Важно оставить асинхронные генерацию и обработку тасков.
// Сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через pull-request в github
// Как видите, никаких привязок к внешним сервисам нет - полный карт-бланш на модификацию кода.

func main() {
	superChan := make(chan Ttype, 10)

	go createTask(superChan)

	doneTasks := make(chan Ttype)

	undoneTasks := make(chan error)

	go taskWorker(superChan, doneTasks, undoneTasks)

	go getTask(doneTasks, undoneTasks)

	time.Sleep(10 * time.Second)

	close(superChan)

	time.Sleep(3 * time.Second)

}

func createTask(a chan Ttype) {
	for {
		select {
		case <-time.After(3 * time.Second):
			ft := time.Now().Format(time.RFC3339)
			if time.Now().Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков
				ft = "Some error occured"
			}
			a <- Ttype{cT: ft, id: int(time.Now().Unix())} // передаем таск на выполнение
		case <-a:
			return
		}
	}
}

func taskWorker(taskChan chan Ttype, tasks chan Ttype, undoneTasks chan error) {
	for task := range taskChan {
		go func(t Ttype) {
			tt, _ := time.Parse(time.RFC3339, t.cT)
			if tt.After(time.Now().Add(-20 * time.Second)) {
				t.taskRESULT = []byte("task has been successed")
			} else {
				t.taskRESULT = []byte("something went wrong")
			}
			t.fT = time.Now().Format(time.RFC3339Nano)

			time.Sleep(time.Millisecond * 150)

			if string(t.taskRESULT[14:]) == "successed" {
				tasks <- t
			} else {
				undoneTasks <- fmt.Errorf("Task id %d time %s, error %s", t.id, t.cT, t.taskRESULT)
			}
		}(task)
	}
}

func getTask(doneTasks chan Ttype, undoneTasks chan error) {
	for {
		select {
		case task, ok := <-doneTasks:
			if ok {
				fmt.Printf("Done task: ID: %d, Time: %sn", task.id, task.cT)
			} else {
				fmt.Println("Done tasks channel closed")
				return
			}
		case err, ok := <-undoneTasks:
			if ok {
				fmt.Printf("Error: %vn", err)
			} else {
				fmt.Println("Undone tasks channel closed")
				return
			}
		case <-time.After(3 * time.Second):
			fmt.Println("----------")
		}
	}
}

// A Ttype represents a meaninglessness of our life
type Ttype struct {
	id         int
	cT         string // время создания
	fT         string // время выполнения
	taskRESULT []byte
}
