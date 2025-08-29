# Staticlint - Инструмент статического анализа Multichecker

## Обзор

Staticlint - это инструмент статического анализа для кода Go, который объединяет несколько анализаторов из различных источников для обеспечения тщательных проверок качества кода. Он включает стандартные анализаторы Go, анализаторы Staticcheck и пользовательские анализаторы для обеспечения высокого качества кода и соблюдения лучших практик.

## Возможности

### Категории анализаторов

#### 1. Стандартные анализаторы Go (golang.org/x/tools/go/analysis/passes)

Это официальные статические анализаторы Go, которые проверяют общие проблемы:

- **asmdecl**: Check assembly declarations
- **assign**: Check for useless assignments
- **atomic**: Check for common mistakes using sync/atomic
- **atomicalign**: Check for non-64-bit-aligned arguments to sync/atomic functions
- **bools**: Check for common mistakes involving boolean operators
- **buildtag**: Check that +build tags are well-formed and correctly located
- **cgocall**: Detect some violations of the cgo pointer passing rules
- **composite**: Check for unkeyed composite literals
- **copylock**: Check for locks erroneously passed by value
- **ctrlflow**: Build a control-flow graph of functions and report unreachable code
- **deepequalerrors**: Check for calls of reflect.DeepEqual on error values
- **errorsas**: Report passing non-pointer or non-error values to errors.As
- **fieldalignment**: Find structs that would use less memory if their fields were sorted
- **findcall**: Find calls to a particular function
- **framepointer**: Report assembly that clobbers the frame pointer before saving it
- **httpresponse**: Check for mistakes using HTTP responses
- **ifaceassert**: Detect impossible interface-to-interface type assertions
- **loopclosure**: Check references to loop variables from within nested functions
- **lostcancel**: Check for failure to call a context cancellation function
- **nilfunc**: Check for useless comparisons between functions and nil
- **nilness**: Check for redundant or impossible nil comparisons
- **pkgfact**: Gather name/value pairs from constant declarations
- **printf**: Check consistency of Printf format strings and arguments
- **reflectvaluecompare**: Check for comparing reflect.Value values with == or reflect.DeepEqual
- **shadow**: Check for possible unintended shadowing of variables
- **shift**: Check for shifts that equal or exceed the width of the integer
- **sigchanyzer**: Detect misuse of unbuffered signal as argument to signal.Notify
- **sortslice**: Check the argument type and order of sort.Slice calls
- **stdmethods**: Check signature of methods of well-known interfaces
- **stringintconv**: Check for string(int) conversions
- **structtag**: Check that struct field tags conform to reflect.StructTag.Get
- **testinggoroutine**: Report calls to (*testing.T).Fatal from goroutines started by a test
- **tests**: Check for common mistaken usages of tests and examples
- **timeformat**: Check for calls of (time.Time).Format or time.Parse with the wrong argument
- **unmarshal**: Report passing non-pointer or non-interface values to unmarshal
- **unreachable**: Check for unreachable code
- **unsafeptr**: Check for invalid uses of unsafe.Pointer
- **unusedresult**: Check for unused results of calls to some functions
- **unusedwrite**: Checks for writes to struct fields and arrays that are never read

#### 2. Анализаторы Staticcheck (класс SA)

Анализаторы Staticcheck SA фокусируются на проблемах корректности:

- **SA1000**: Invalid regular expression
- **SA1001**: Invalid template
- **SA1002**: Invalid time format
- **SA1003**: Invalid argument to encoding/binary
- **SA1004**: Suspiciously small untyped constant in time.Sleep
- **SA1005**: Invalid first argument to exec.Command
- **SA1006**: Printf with dynamic first argument and no further arguments
- **SA1007**: Invalid URL in net/url.Parse
- **SA1008**: Non-canonical key in http.Header map
- **SA1009**: Invalid argument for append
- **SA1010**: (*regexp.Regexp).FindAll called with n == 0, which will always return zero results
- **SA1011**: Various methods in the strings package expect valid UTF-8, but invalid UTF-8 is given
- **SA1012**: A nil context.Context is being passed to a function, consider using context.TODO instead
- **SA1013**: io.Seeker.Seek is being called with the whence constant as the first argument, but it should be the second
- **SA1014**: Non-pointer value passed to Unmarshal or Decode
- **SA1015**: Using time.Tick in a way that will leak. Consider using time.NewTicker, and only use time.Tick in tests, commands and endless functions
- **SA1016**: Trapping a signal that cannot be trapped
- **SA1017**: Channels used with os/signal.Notify should be buffered
- **SA1018**: strings.Replace called with n == 0, which does nothing
- **SA1019**: Using a deprecated function, variable, constant or field
- **SA1020**: Using an invalid host:port pair with a net.Listen-related function
- **SA1021**: Using bytes.Equal to compare two net.IP
- **SA1023**: Modifying the buffer in an io.Writer implementation
- **SA1024**: A string cutset contains duplicate characters
- **SA1025**: It is not possible to use (*time.Timer).Reset's return value correctly
- **SA1026**: Cannot marshal channels or functions
- **SA1027**: Atomic access to 64-bit variable must be 64-bit aligned
- **SA1028**: sort.Slice can only be used on slices
- **SA1029**: Inappropriate key in call to context.WithValue
- **SA1030**: Invalid argument in call to a strconv function
- **SA2000**: sync.WaitGroup.Add called with negative delta
- **SA2001**: Empty critical section, did you mean to defer the unlock?
- **SA2002**: Called testing.T.FailNow or SkipNow in a goroutine, which isn't allowed
- **SA2003**: Deferred Lock right after Lock, probably meant to defer Unlock
- **SA3000**: Test functions shouldn't return anything, have a return type
- **SA3001**: Assigning to b.N in benchmarks is unsupported
- **SA4000**: Boolean expression has identical expressions on both sides
- **SA4001**: &*x gets simplified to x, it does not copy x
- **SA4002**: Comparing to boolean constant
- **SA4003**: Comparing unsigned variable against negative number
- **SA4004**: The loop exits unconditionally after one iteration
- **SA4005**: Field assignment that will never be observed. Did you mean to use a pointer receiver?
- **SA4006**: A value assigned to a variable is never read before being overwritten. Forgotten error check or dead code?
- **SA4008**: The variable in the loop condition never changes, are you incrementing the wrong variable?
- **SA4009**: A function argument is overwritten before its first use
- **SA4010**: The result of append will never be observed anywhere
- **SA4011**: Break statement with no effect. Did you mean to break out of an outer loop?
- **SA4012**: Comparing a value against NaN even though no value is equal to NaN
- **SA4013**: Negating a boolean twice (!!b) is the same as writing b. This is either redundant, or a typo
- **SA4014**: An if/else if chain has repeated conditions and no side-effects; if the condition didn't match the first time, it won't match the second time, either
- **SA4015**: Calling math.IsNaN on a value that cannot be a NaN
- **SA4016**: Certain bitwise operations, such as x ^ 0, can be simplified to x
- **SA4017**: A pure function's return value is discarded, making the call pointless
- **SA4018**: Self-assignment of variables
- **SA4019**: Multiple, equally valid constant definitions of the same value under different names
- **SA4020**: Unreachable case clause in a type switch
- **SA4021**: x = append(y) is equivalent to x = y
- **SA4022**: Comparing the address of a variable against nil
- **SA4023**: Impossible comparison of interface value with untyped nil
- **SA5000**: Assignment to nil map
- **SA5001**: Defering Close before checking for a possible error
- **SA5002**: The empty for loop (for {}) spins and can block the scheduler
- **SA5003**: Defers in infinite loops will never execute
- **SA5004**: for loop selects the default case
- **SA5005**: The finalizer references the finalized object, preventing garbage collection
- **SA5006**: Setting field with pointer type to zero value
- **SA5007**: Infinite recursive call
- **SA5008**: Invalid struct tag
- **SA5009**: Invalid Printf call
- **SA5010**: Impossible type assertion
- **SA5011**: Possible nil pointer dereference
- **SA5012**: Passing odd-sized slice to crypto.rc4
- **SA6000**: Using regexp.Match or regexp.MatchString with context that should use regexp.Compile
- **SA6001**: Missing an optimization opportunity when indexing maps by byte slices
- **SA6002**: Store to nil map
- **SA6003**: Converting a string to a slice of runes before ranging over it
- **SA6004**: Inefficient string comparison with strings.ToLower or strings.ToUpper
- **SA6005**: Inefficient string comparison with strings.TrimPrefix or strings.TrimSuffix
- **SA9001**: Defers in range loops may not execute when you expect them to
- **SA9002**: Using a non-octal os.FileMode that looks like it was meant to be in octal
- **SA9003**: Empty body in an if or else branch
- **SA9004**: Only the first constant has an explicit type
- **SA9005**: Trying to marshal a struct with no public fields nor custom marshaling
- **SA9006**: Dubious bit shifting of a fixed size integer value
- **SA9007**: Deleting a directory that shouldn't be deleted
- **SA9008**: else branch of if statement ends with a return statement, so drop this else and outdent its block
- **SA9009**: Missing calls to error check
- **SA9010**: Using reflect.TypeOf as the key in a map
- **SA9011**: Using deprecated driver.Conn interface
- **SA9012**: A function's default error return value is unused

#### 3. Анализаторы Stylecheck (класс ST)

Анализаторы Stylecheck фокусируются на стиле кода и форматировании:

- **ST1000**: Incorrect or missing package comment
- **ST1001**: Dot imports are discouraged
- **ST1003**: Poorly chosen identifier
- **ST1005**: Incorrectly formatted error string
- **ST1006**: Poorly chosen receiver name
- **ST1008**: A function's error value should be its last return value
- **ST1011**: Poorly chosen name for variable of type time.Duration
- **ST1012**: Poorly chosen name for error variable
- **ST1013**: Should use constants for HTTP error codes, not magic numbers
- **ST1015**: A switch case clause can be simplified to a case statement
- **ST1016**: Use consistent method receiver names
- **ST1017**: Don't use Yoda conditions
- **ST1018**: Avoid zero-width and control characters in string literals
- **ST1019**: Importing the same package multiple times
- **ST1020**: The documentation of an exported function should start with the function's name
- **ST1021**: The documentation of an exported type should start with type's name
- **ST1022**: The documentation of an exported variable or constant should start with variable's name
- **ST1023**: Redundant type in variable declaration

#### 4. Простые анализаторы (класс S)

Простые анализаторы фокусируются на упрощении кода:

- **S1000**: Use plain channel send or receive instead of single-case select
- **S1001**: Replace for loop with call to copy
- **S1002**: Omit comparison with boolean constant
- **S1003**: Replace call to strings.Index with strings.Contains
- **S1004**: Replace call to bytes.Compare with bytes.Equal
- **S1005**: Drop unnecessary use of the blank identifier
- **S1006**: Use for range instead of for loop
- **S1007**: Simplify regular expression by using raw string literal
- **S1008**: Simplify returning boolean expression
- **S1009**: Omit redundant nil check on slices
- **S1010**: Omit default slice index
- **S1011**: Use a single append to concatenate two slices
- **S1012**: Replace time.Now().Sub(x) with time.Since(x)
- **S1016**: Use a type conversion instead of struct literal
- **S1017**: Replace manual trimming with strings.TrimPrefix
- **S1018**: Use copy for sliding elements
- **S1019**: Simplify make call by omitting redundant arguments
- **S1020**: Omit redundant nil check in type assertion
- **S1021**: Merge variable declaration and assignment
- **S1023**: Omit redundant control flow
- **S1024**: Replace x.Sub(time.Now()) with time.Until(x)
- **S1025**: Don't use fmt.Sprintf("%s", x) unnecessarily
- **S1028**: Simplify error construction with fmt.Errorf
- **S1029**: Range over the string directly
- **S1030**: Use bytes.Buffer.String or bytes.Buffer.Bytes
- **S1031**: Omit redundant nil check around loop
- **S1032**: Use sort.Ints(x), sort.Float64s(x), and sort.Strings(x)
- **S1033**: Unnecessary guard around call to delete
- **S1034**: Assigning the result of this guard statement to a variable
- **S1035**: Redundant call to net/http.CanonicalHeaderKey in method call on http.Header
- **S1036**: Unnecessary guard around map access
- **S1037**: Unnecessary guard around channel receive
- **S1038**: Unnecessary use of fmt.Sprint
- **S1039**: Unnecessary use of fmt.Sprint
- **S1040**: Type assertion to current type

#### 5. Пользовательские анализаторы

##### Анализатор Exitcheck

Пользовательский анализатор `exitcheck` запрещает использование вызовов `os.Exit` в пакете main. Это помогает обеспечить лучшие практики обработки ошибок:

- Предотвращение прямых вызовов `os.Exit()` в функциях пакета main
- Обнаружение вызовов `log.Fatal()` и `log.Fatalf()`, которые внутри вызывают `os.Exit()`
- Правильная обработка ошибок и механизмов корректного завершения

**Примеры нарушений:**
```go
func main() {
    if err != nil {
        os.Exit(1) // ❌ Это будет ошибка
    }
    log.Fatal("error") // ❌ Это тоже будет ошибка
}
```

**Рекомендуемые альтернативы:**
```go
func main() {
    if err != nil {
        log.Printf("error: %v", err)
        return // или используйте правильный механизм обработки ошибок
    }
}
```

## Установка

Инструмент staticlint является частью проекта shortlink. Использование:

```bash
# Перейдите в директорию проекта
cd shortlink

# Запустите инструмент staticlint
go run cmd/staticlint/main.go ./...
```

## Использование

### Базовое использование

```bash
# Анализ всех пакетов в текущей директории и поддиректориях
go run cmd/staticlint/main.go ./...

# Анализ конкретных пакетов
go run cmd/staticlint/main.go ./cmd/... ./internal/...

# Анализ одного файла
go run cmd/staticlint/main.go ./cmd/shortener/main.go
```

### Формат вывода

Инструмент выводит сообщения в следующем формате:
```
filename:line:column: analyzer_name: message
```

Пример:
```
cmd/config/config.go:48:9: exitcheck: call to log.Fatal in main package (calls os.Exit)
```

## Конфигурация

Инструмент staticlint использует конфигурацию по умолчанию для всех анализаторов. Для настройки поведения анализаторов вы можете изменить настройки анализаторов в файле `main.go`.

## Устранение неполадок

### Общие проблемы

1. **Import errors**: Ensure all required dependencies are installed
2. **False positives**: Some analyzers may report false positives. Review each case individually
3. **Performance**: For large codebases, consider running analyzers selectively

### Советы по производительности

- Запускайте анализаторы на конкретных пакетах, а не на всей кодовой базе
- Используйте `go build` перед запуском статического анализа, чтобы убедиться, что код компилируется
- Рассмотрите возможность отдельного запуска тяжелых анализаторов (например, fieldalignment)

## Лицензия

Этот инструмент является частью проекта shortlink и следует тем же условиям лицензии.
