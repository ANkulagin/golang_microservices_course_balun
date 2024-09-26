/*
Этот файл представляет собой клиентское приложение для взаимодействия с HTTP сервером, реализованным в предыдущем файле.
Он предоставляет функции для создания новой заметки на сервере и получения существующей заметки по её ID. В качестве данных для заметок используются сгенерированные случайные значения с помощью библиотеки gofakeit.

Связь с предыдущим файлом:
- Сервер из предыдущего файла предоставляет API для создания и получения заметок.
- Этот клиент отправляет запросы на сервер для создания и получения заметок, тестируя работу API.

Основные функции:
- createNoteClient: Отправляет POST-запрос на сервер для создания новой заметки.
- getNoteClient: Отправляет GET-запрос на сервер для получения заметки по её ID.
*/

package main

import (
	"bytes"         // Пакет для работы с буферами в памяти
	"encoding/json" // Пакет для кодирования и декодирования JSON
	"fmt"           // Пакет для форматирования строк
	"io"            // Пакет для работы с потоками ввода/вывода
	"log"           // Пакет для логирования
	"net/http"      // Пакет для отправки HTTP запросов

	"github.com/brianvoe/gofakeit" // Библиотека для генерации случайных данных
	"github.com/fatih/color"       // Библиотека для цветного вывода в консоль
	"github.com/pkg/errors"        // Пакет для работы с ошибками
)

const (
	// baseUrl определяет базовый URL для обращения к серверу
	baseUrl = "http://localhost:8081"
	// createPostfix определяет путь для создания новой заметки на сервере
	createPostfix = "/notes"
	// getPostfix определяет путь для получения заметки по её ID. %d - формат для числового ID
	getPostfix = "/notes/%d"
)

// NoteInfo содержит информацию о заметке, аналогично структуре из серверного кода
type NoteInfo struct {
	Title    string `json:"title"`     // Заголовок заметки
	Context  string `json:"context"`   // Содержание заметки
	Author   string `json:"author"`    // Автор заметки
	IsPublic bool   `json:"is_public"` // Флаг, указывающий, является ли заметка публичной
}

// Note представляет заметку с уникальным ID и временными метками, аналогично серверной структуре
type Note struct {
	ID        int64    `json:"id"`         // Уникальный идентификатор заметки
	Info      NoteInfo `json:"info"`       // Вложенная структура с информацией о заметке
	CreatedAt string   `json:"created_at"` // Временная метка создания заметки
	UpdatedAt string   `json:"updated_at"` // Временная метка последнего обновления заметки
}

// createNoteClient создает новую заметку, отправляя POST-запрос на сервер
// Возвращает созданную заметку и ошибку, если что-то пошло не так
func createNoteClient() (Note, error) {
	// Генерируем случайные данные для заметки с помощью gofakeit
	note := NoteInfo{
		Title:    gofakeit.BeerName(),    // Генерация случайного названия для заголовка
		Context:  gofakeit.IPv4Address(), // Генерация случайного IP адреса для содержания (для примера)
		Author:   gofakeit.Name(),        // Генерация случайного имени автора
		IsPublic: gofakeit.Bool(),        // Генерация случайного булева значения для публичности заметки
	}

	// Сериализуем структуру NoteInfo в JSON для отправки на сервер
	data, err := json.Marshal(note)
	if err != nil {
		// Возвращаем пустую заметку и ошибку, если сериализация не удалась
		return Note{}, err
	}

	// Отправляем POST-запрос на сервер с JSON данными
	resp, err := http.Post(baseUrl+createPostfix, "application/json", bytes.NewBuffer(data))
	if err != nil {
		// Возвращаем пустую заметку и ошибку, если запрос не удался
		return Note{}, err
	}
	// Гарантируем закрытие тела ответа после завершения функции
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			// Логируем ошибку, если не удалось закрыть тело ответа
			log.Fatal("Failed to close body:", err)
		}
	}(resp.Body)

	// Если сервер вернул статус, отличный от 201 (Created), возвращаем ошибку
	if resp.StatusCode != http.StatusCreated {
		return Note{}, err
	}

	// Декодируем JSON ответ от сервера в структуру Note
	var createdNote Note
	if err = json.NewDecoder(resp.Body).Decode(&createdNote); err != nil {
		// Возвращаем пустую заметку и ошибку, если декодирование не удалось
		return Note{}, err
	}

	// Возвращаем созданную заметку и nil, если все прошло успешно
	return createdNote, nil
}

// getNoteClient получает заметку по её ID, отправляя GET-запрос на сервер
// Возвращает найденную заметку и ошибку, если что-то пошло не так
func getNoteClient(id int64) (Note, error) {
	// Форматируем URL с ID заметки и отправляем GET-запрос на сервер
	resp, err := http.Get(fmt.Sprintf(baseUrl+getPostfix, id))
	if err != nil {
		// Логируем и завершаем программу, если запрос не удался
		log.Fatal("Failed to get note:", err)
	}
	// Гарантируем закрытие тела ответа после завершения функции
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			// Логируем ошибку, если не удалось закрыть тело ответа
			log.Fatal("Failed to close body:", err)
		}
	}(resp.Body)

	// Если сервер вернул статус 404 (Not Found), возвращаем пустую заметку и ошибку
	if resp.StatusCode == http.StatusNotFound {
		return Note{}, err
	}

	// Если сервер вернул статус, отличный от 200 (OK), возвращаем ошибку
	if resp.StatusCode != http.StatusOK {
		return Note{}, errors.Errorf("failed to get note: %d", resp.StatusCode)
	}

	// Декодируем JSON ответ от сервера в структуру Note
	var note Note
	if err = json.NewDecoder(resp.Body).Decode(&note); err != nil {
		// Возвращаем пустую заметку и ошибку, если декодирование не удалось
		return Note{}, err
	}

	// Возвращаем найденную заметку и nil, если все прошло успешно
	return note, nil
}

// main является точкой входа в клиентское приложение
// Оно создает заметку на сервере и затем пытается получить её по ID, логируя результаты
func main() {
	// Создаем новую заметку на сервере
	note, err := createNoteClient()
	if err != nil {
		// Логируем и завершаем программу, если создание заметки не удалось
		log.Fatal("failed to create note:", err)
	}

	// Логируем информацию о созданной заметке, выводя её в цвете
	log.Printf(color.RedString("Note created:\n"), color.GreenString("%#+v", note))

	// Пытаемся получить заметку по её ID с сервера
	note, err = getNoteClient(note.ID)
	if err != nil {
		// Логируем и завершаем программу, если получение заметки не удалось
		log.Fatal("failed to get note:", err)
	}

	// Логируем информацию о полученной заметке, выводя её в цвете
	log.Printf(color.RedString("Note info got:\n"), color.GreenString("%#+v", note))
}
