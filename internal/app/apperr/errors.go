package apperr

import "errors"

var (
	// ErrLinkExists - попытка создать короткую ссылку для существующего оригинального URL.
	ErrLinkExists = errors.New("link already exists")

	// ErrLinkNotFound - запрошенная короткая ссылка отсутствует.
	ErrLinkNotFound = errors.New("link not found")

	// ErrBodyRead -  ошибка чтении тела запроса.
	ErrBodyRead = errors.New("cannot read the body")

	// ErrOnlyGET - эндпоинт принимает только запросы GET.
	ErrOnlyGET = errors.New("only GET requests are allowed")

	// ErrOnlyPOST - эндпоинт принимает только запросы POST.
	ErrOnlyPOST = errors.New("only POST requests are allowed")

	// ErrValAlreadyExists - конфликт, связанный с существованием значения.
	ErrValAlreadyExists = errors.New("conflict")
)
