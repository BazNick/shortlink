package apperr

import "errors"

var (
	// ErrLinkExists - попытка создать короткую ссылку для существующего оригинального URL.
	ErrLinkExists = errors.New("ссылка уже существует")

	// ErrLinkNotFound - запрошенная короткая ссылка отсутствует.
	ErrLinkNotFound = errors.New("ссылка не найдена")

	// ErrBodyRead -  ошибка чтении тела запроса.
	ErrBodyRead = errors.New("не удается прочитать тело запроса")

	// ErrOnlyGET - эндпоинт принимает только запросы GET.
	ErrOnlyGET = errors.New("разрешены только запросы GET")

	// ErrOnlyPOST - эндпоинт принимает только запросы POST.
	ErrOnlyPOST = errors.New("разрешены только запросы POST")

	// ErrValAlreadyExists - конфликт, связанный с существованием значения.
	ErrValAlreadyExists = errors.New("конфликт")
)
