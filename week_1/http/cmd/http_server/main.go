/*
Этот файл содержит реализацию простого HTTP сервера на Go, который использует библиотеку Chi для маршрутизации запросов.
Сервер предоставляет два маршрута:
- POST /notes — для создания новой заметки. Обработчик createNoteHandler принимает информацию о заметке в формате JSON, сохраняет её и возвращает созданную заметку с уникальным ID.
- GET /notes/{id} — для получения заметки по ID. Обработчик getNoteHandler ищет заметку по её ID и возвращает её, если она существует.

Для хранения заметок используется потокобезопасная структура SyncMap, которая обеспечивает безопасный доступ к заметкам из разных потоков.
*/

package main

import (
	"encoding/json"            // Пакет для работы с JSON: сериализация и десериализация данных
	"github.com/go-chi/chi/v5" // Пакет для маршрутизации HTTP запросов
	"log"                      // Пакет для логирования
	"math/rand"                // Пакет для генерации случайных чисел
	"net/http"                 // Пакет для работы с HTTP запросами и ответами
	"strconv"                  // Пакет для преобразования строк в числа и наоборот
	"sync"                     // Пакет для работы с примитивами синхронизации, такими как мьютексы
	"time"                     // Пакет для работы со временем и временными метками
)

const (
	// baseUrl определяет базовый адрес, на котором будет запущен сервер
	baseUrl = "localhost:8081"
	// createPostfix определяет путь для создания новой заметки
	createPostfix = "/notes"
	// getPostfix определяет путь для получения заметки по её ID. %d - это формат для числового ID.
	getPostfix = "/notes/%d"
)

// NoteInfo содержит информацию о заметке
type NoteInfo struct {
	Title    string `json:"title"`     // Заголовок заметки
	Context  string `json:"context"`   // Основное содержание заметки
	Author   string `json:"author"`    // Имя автора заметки
	IsPublic bool   `json:"is_public"` // Флаг, указывающий, является ли заметка публичной
}

// Note представляет заметку с уникальным ID, информацией и временными метками
type Note struct {
	ID        int64     `json:"id"`         // Уникальный идентификатор заметки
	Info      NoteInfo  `json:"info"`       // Вложенная структура, содержащая информацию о заметке
	CreatedAt time.Time `json:"created_at"` // Временная метка создания заметки
	UpdatedAt time.Time `json:"updated_at"` // Временная метка последнего обновления заметки
}

// SyncMap представляет потокобезопасную карту для хранения заметок
type SyncMap struct {
	elems map[int64]*Note // Карта для хранения заметок, где ключом является ID заметки
	m     sync.RWMutex    // Мьютекс для синхронизации доступа к карте заметок
}

// Инициализация глобальной переменной notes, которая представляет собой потокобезопасную карту для всех заметок
var notes = &SyncMap{
	elems: make(map[int64]*Note), // Создаем пустую карту для хранения заметок
}

// createNoteHandler обрабатывает HTTP запрос на создание новой заметки
// Он принимает информацию о заметке в формате JSON из тела запроса, создает новую заметку и добавляет её в карту notes
func createNoteHandler(w http.ResponseWriter, r *http.Request) {
	info := &NoteInfo{} // Создаем указатель на структуру NoteInfo для хранения данных заметки

	// Декодируем JSON из тела запроса в структуру NoteInfo
	if err := json.NewDecoder(r.Body).Decode(info); err != nil {
		// Если при декодировании произошла ошибка, отправляем клиенту ответ с кодом 400 (Bad Request)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Инициализируем генератор случайных чисел текущим временем
	rand.Seed(time.Now().UnixNano())
	now := time.Now() // Получаем текущее время для использования в CreatedAt и UpdatedAt

	// Создаем новую заметку с уникальным ID и текущими временными метками
	note := &Note{
		ID:        rand.Int63(), // Генерируем случайный 64-битный ID для заметки
		Info:      *info,        // Копируем данные из info в структуру Note
		CreatedAt: now,          // Устанавливаем время создания заметки
		UpdatedAt: now,          // Устанавливаем время последнего обновления заметки
	}

	// Устанавливаем заголовок Content-Type для ответа как "application/json"
	w.Header().Set("Content-Type", "application/json")
	// Устанавливаем код ответа как 201 (Created), так как заметка успешно создана
	w.WriteHeader(http.StatusCreated)
	// Кодируем структуру Note в JSON и отправляем её в ответе
	if err := json.NewEncoder(w).Encode(note); err != nil {
		// Если произошла ошибка при кодировании JSON, отправляем клиенту ответ с кодом 500 (Internal Server Error)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Блокируем карту заметок для записи
	notes.m.Lock()
	defer notes.m.Unlock() // Разблокируем карту после завершения работы функции

	// Добавляем новую заметку в карту, используя её ID в качестве ключа
	notes.elems[note.ID] = note
}

// getNoteHandler обрабатывает HTTP запрос на получение заметки по её ID
// Он принимает ID заметки из URL, ищет заметку в карте notes и возвращает её клиенту в формате JSON
func getNoteHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем параметр "id" из URL, используя chi.URLParam
	noteID := chi.URLParam(r, "id")
	// Парсим строковый ID в целочисленный тип
	id, err := parseNoteID(noteID)
	if err != nil {
		// Если ID не является допустимым целым числом, отправляем клиенту ответ с кодом 400 (Bad Request)
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	// Блокируем карту заметок для чтения
	notes.m.RLock()
	defer notes.m.RUnlock() // Разблокируем карту после завершения работы функции

	// Ищем заметку в карте по её ID
	note, ok := notes.elems[id]
	if !ok {
		// Если заметка не найдена, отправляем клиенту ответ с кодом 404 (Not Found)
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}

	// Устанавливаем заголовок Content-Type для ответа как "application/json"
	w.Header().Set("Content-Type", "application/json")
	// Кодируем структуру Note в JSON и отправляем её в ответе
	if err = json.NewEncoder(w).Encode(note); err != nil {
		// Если произошла ошибка при кодировании JSON, отправляем клиенту ответ с кодом 500 (Internal Server Error)
		http.Error(w, "Failed to encode note data", http.StatusInternalServerError)
		return
	}
}

// parseNoteID парсит строковый идентификатор в целое число (int64)
// Он принимает строковый ID и возвращает его как int64 или ошибку, если парсинг не удался
func parseNoteID(idStr string) (int64, error) {
	// Используем strconv.ParseInt для преобразования строки в int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// Возвращаем нулевое значение и ошибку, если парсинг не удался
		return 0, err
	}

	// Возвращаем распарсенный ID и nil, если все прошло успешно
	return id, nil
}

// main инициализирует HTTP сервер, настраивает маршруты и запускает сервер
func main() {
	// Создаем новый роутер с использованием библиотеки Chi
	r := chi.NewRouter()

	// Регистрируем маршрут для POST запросов по пути createPostfix, который будет обрабатывать createNoteHandler
	r.Post(createPostfix, createNoteHandler)

	// Регистрируем маршрут для GET запросов по пути getPostfix, который будет обрабатывать getNoteHandler
	r.Get(getPostfix, getNoteHandler)

	// Запускаем HTTP сервер на baseUrl и связываем его с роутером r
	err := http.ListenAndServe(baseUrl, r)
	if err != nil {
		// Логируем ошибку и завершает выполнение программы, если сервер не может быть запущен
		log.Fatal(err)
	}
}
