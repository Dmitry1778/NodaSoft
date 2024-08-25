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
	superChan := make(chan Ttype, 10) // Канал для создания тасков

	go createTask(superChan) // Запускаем горутину для генерации тасков

	doneTasks := make(chan Ttype) // Канал для завершенных тасков

	undoneTasks := make(chan error) // Канал для ошибок

	go taskWorker(superChan, doneTasks, undoneTasks) // Запускаем горутину для обработки тасков

	go getTask(doneTasks, undoneTasks) // Запускаем горутину для вывода результатов

	time.Sleep(10 * time.Second)

	close(superChan) //Закрытие канала для создания тасков

	time.Sleep(3 * time.Second)

}

// Методот генерации тасков
func createTask(taskChan chan Ttype) {
	for {
		select {
		case <-time.After(3 * time.Second): // Каждые 3 секунды генерации
			ft := time.Now().Format(time.RFC3339)
			if time.Now().Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков
				ft = "Some error occured"
			}
			taskChan <- Ttype{cT: ft, id: int(time.Now().Unix())} // передаем таск на выполнение
		case <-taskChan:
			return
		}
	}
}

// Метод обработки тасков
func taskWorker(taskChan chan Ttype, tasks chan Ttype, undoneTasks chan error) {
	for task := range taskChan {
		go func(t Ttype) {
			tt, _ := time.Parse(time.RFC3339, t.cT)          // Извлекаем время из t.cT
			if tt.After(time.Now().Add(-20 * time.Second)) { // Проверяем, было ли время создания задачи менее 20 секунд назад
				t.taskRESULT = []byte("task has been successed") // Если да то записываем такое результат
			} else {
				t.taskRESULT = []byte("something went wrong") // Иначе такое
			}
			t.fT = time.Now().Format(time.RFC3339Nano) // В поле t.fT записывается текущее время завершения задачи

			time.Sleep(time.Millisecond * 150) // Искусственная задержка в 150 миллисекунд для имитации обработки задачи

			if string(t.taskRESULT[14:]) == "successed" { // Проверка результата taskResult
				tasks <- t // записываем в tasks
			} else {
				undoneTasks <- fmt.Errorf("Task id %d time %s, error %s", t.id, t.cT, t.taskRESULT) // Иначе записываем ошибку
			}
		}(task) // Пробегаемся по всем таскам через анонимную функцию. Количество пробежки зависит от количества tasks
	}
}

// Метод получения tasks
func getTask(doneTasks chan Ttype, undoneTasks chan error) {
	for { // Бесконечный цикл выполняется пока не функция не завершится
		select { // ожидает, пока один из каналов будет готов
		case task, ok := <-doneTasks:
			if ok {
				fmt.Printf("Done task: ID: %d, Time: %sn", task.id, task.cT) // Если есть таск выводим это
			} else {
				fmt.Println("Done tasks channel closed")
				return
			}
		case err, ok := <-undoneTasks:
			if ok {
				fmt.Printf("Error: %vn", err) // Если есть нет таск выводим это
			} else {
				fmt.Println("Undone tasks channel closed")
				return
			}
		case <-time.After(3 * time.Second): // Если ни один из каналов не был готов в течение 3 секунд, выводится строка "----------"
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
