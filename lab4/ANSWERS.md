# Lab 4 — ответы на вопросы

Запуск: `go run ./task1` … `go run ./task7`.
Бенчмарк: `go test -run='^$' -bench=. -benchmem ./task6`.

## Part 1: Interface semantics

### Что внутри интерфейса: `%v %T`
```
<nil> <nil>          // var s Shape — пустой интерфейс: ни типа, ни значения
{1 2} main.Rect      // s = Rect{1, 2}
&{2} *main.Circle    // s = &Circle{2}
```
`%T` печатает **динамический тип**, то есть тип значения, которое сейчас лежит внутри, а не `Shape`. `Shape` — это статический тип переменной, его знает только компилятор. При передаче в `Printf(format string, a ...any)` интерфейс `Shape` превращается в `any`, и переносится пара *(динамический тип, значение)*. У nil-интерфейса динамического типа нет, поэтому `%T` даёт `<nil>`.

### Task 1: Notifications — [task1/main.go](task1/main.go)
- `Email` и `SMS` ничего не меняют, поэтому у них value receiver.
- `SMS` считает **символы**, то есть руны (`utf8.RuneCountInString`), а не байты (`len`). Иначе сообщение из 11 кириллических букв (22 байта) было бы отклонено (Week 3). Ошибка оборачивает sentinel `ErrTooLong` через `%w`, поэтому работает `errors.Is(err, ErrTooLong)`.
- `Telegram` меняет `Sent`, поэтому нужен **pointer receiver**. С value receiver `Sent++` увеличил бы копию, и счётчик остался бы 0. Следствие (Week 4, method sets): `Notifier` реализует только `*Telegram`, поэтому в слайс кладём `tg := &Telegram{...}`.
  ```
  var _ Notifier = Telegram{}
  // cannot use Telegram{} (value of struct type Telegram) as Notifier value in variable declaration:
  // Telegram does not implement Notifier (method Notify has pointer receiver)
  ```
- Compile-time проверки: `var _ Notifier = Email{}`, `SMS{}`, `(*Telegram)(nil)`. `(*Telegram)(nil)` — nil-указатель нужного типа: ничего не аллоцирует, но тип у него есть, и компилятор проверяет его method set.
- `Broadcast` начинает с `var errs []error`, поэтому, если ошибок нет, возвращает `nil`.

```
email    -> aruzhan@uni.kz: Lab 4 is out
sms      -> +77001234567: Lab 4 is out
telegram -> @se-2401: Lab 4 is out
email    -> aruzhan@uni.kz: Defence is moved to Friday, room 305
telegram -> @se-2401: Defence is moved to Friday, room 305
  error: sms to +77001234567: 36 characters, max 20: message too long
  is ErrTooLong: true
telegram sent: 2
```

**Participation task:** `printGrades(w io.Writer, grades []Grade)` вызывается с `os.Stdout` (`*os.File`) и с `&sb`. Именно `&sb`, потому что `Write` у `strings.Builder` объявлен на указателе. Функция не знает, куда пишет: в терминал или в строку.

## Part 2: Type assertions and type switches

### Single-value assertion
```
Go
0 false
panic: interface conversion: interface {} is string, not float64
```

### Почему `case int, int64: return x + 1` не компилируется?
```
invalid operation: x + 1 (mismatched types any and untyped int)
```
Переменная в Go имеет **один** статический тип. В case с несколькими типами компилятор не может выбрать между `int` и `int64`, поэтому `x` остаётся `any`, а у `any` нет операции `+`. Исправление — отдельные case: `case int: return x + 1` и `case int64: return x + 1`.

### Timeout
```
Stringer            // kind(Timeout{5})
timeout after 3s    // fmt.Println(e)
```
- `Timeout` реализует **и** `fmt.Stringer`, **и** `error`. Case проверяются сверху вниз, побеждает первый совпавший, то есть `fmt.Stringer`. Если поменять первые два case местами, будет напечатано `error`. До `case Timeout` дело не дойдёт никогда, а компилятор об этом не предупреждает.
- `fmt.Println(e)` вызывает `Error()`: внутри `fmt` (`handleMethods`) сначала проверяется `case error:`, потом `case Stringer:`. Статический тип `error` у переменной `e` здесь ни при чём: в `Println` она приходит как `any`.

### Task 2: Describe anything — [task2/main.go](task2/main.go)

| Значение | Результат | Почему (типичные ошибки в предсказании) |
|---|---|---|
| `nil` | `nil` | `case nil` срабатывает только для nil-**интерфейса** |
| `7` | `integer 7` | |
| `int64(7)` | `integer 7` | тот же case `int, int64` |
| `"привет"` | `string of 12 bytes` | `len` считает байты; кириллица в UTF-8 = 2 байта на руну, 6 × 2 = 12 |
| `[]int{1, 2, 3}` | `slice with 3 ints` | |
| `errors.New("disk full")` | `error: disk full` | `*errors.errorString` реализует `error`; case `error` стоит раньше `Stringer` |
| `Celsius(21.5)` | `Stringer: 21.5°C` | `Celsius` — **отдельный** именованный тип, не `float64`. Совпал по методу `String()` |
| `3.14` | `unknown type float64` | нетипизированная константа получает тип по умолчанию `float64`, а case для него нет |
| `struct{}{}` | `unknown type struct {}` | `%T` печатает анонимный тип как `struct {}` |
| `true` | `bool true` | новый case |
| `map[string]int{...}` | `map with 2 keys` | новый case |

Программа сама сверяет результат с ожиданием и с `describeAssert`: все строки `ok`.

**Participation task:** `describeAssert` (только `v, ok := i.(T)`) работает так же, но читается хуже. На каждый тип уходит отдельный `if` с повтором `v.(T)`, а `case int, int64` превращается в два `if`. В type switch переменная `x` уже нужного типа, вся логика — одна конструкция, а компилятор ловит дубли (`duplicate case int in type switch`). Type switch проще.

## Part 3: Nil interfaces

### Square
```
true                         // var s Shape — (nil, nil)
false                        // s = sq — (*Square, nil)
0                            // s.Area() — метод вызван с nil-ресивером
panic: runtime error: invalid memory address or nil pointer dereference
```
Когда в интерфейсе лежит `(*Square, nil)`, **тип известен**. Go находит метод `(*Square).Area` в таблице методов (itab) и вызывает его, передав `nil` как ресивер, а метод сам обрабатывает `s == nil`. Когда интерфейс целиком `nil`, типа нет, искать метод негде: паника.

### Task 3: Find the bug — [task3/main.go](task3/main.go)
```
--- buggy
error: <nil>
err == nil: false
%T: *main.ValidationError
%v: <nil>
pointer inside == nil: true
--- fixed
validate(20) == nil: true
%T: <nil>
validate(-1): field = age
wrapped: register user: invalid field: age
  assertion ok: false
  errors.As field = age
```
1. Печатается `error: <nil>`. `return err` превращает `*ValidationError(nil)` в интерфейс `error` = `(*ValidationError, nil)`. Тип есть, значит интерфейс **не** nil, и `if` выполняется.
2. Почему `<nil>`: `fmt.Println` вызывает `err.Error()`, а там `e.Field` разыменовывает nil-указатель, то есть паника. `fmt` ловит её (`catchPanic`): если аргумент — nil-указатель, печатает `<nil>` вместо падения программы.
3. Доказательство: `err == nil: false`, а `%T` = `*main.ValidationError`.
4. Исправление: тип результата `error`, и для «нет ошибки» пишем литерал `return nil`.
5. `errors.As(err, &ve)`: второй аргумент — указатель на переменную того типа, который ищем (`*ValidationError`). В отличие от `err.(*ValidationError)`, `errors.As` разворачивает цепочку `%w`. Пример с `wrapped`: assertion дал `false`, а `errors.As` нашёл `Field`.

### Task 4: Exercise: Errors — [task4/main.go](task4/main.go)
```
1.4142135623730951 <nil>
0 cannot Sqrt negative number: -2
0 <nil>
12 <nil>
main.ErrNegativeSqrt: cannot Sqrt negative number: -9
```
**Почему `fmt.Sprint(e)` внутри `Error()` даёт бесконечный цикл?** `e` имеет тип `ErrNegativeSqrt`, а он реализует `error`. `fmt` для такого аргумента вызывает `e.Error()`, который снова вызывает `fmt.Sprint(e)`, и так далее. Рекурсия без выхода:
```
runtime: goroutine stack exceeds 1000000000-byte limit
fatal error: stack overflow
```
Исправление: `float64(e)`. У `float64` нет методов, поэтому `fmt` печатает просто число.

**Как это ловит `go vet`:** анализатор `printf` знает функции `fmt`. Он видит, что внутри метода `Error()` (или `String()`) в `fmt.Sprint`/`Sprintf` передаётся сам ресивер, тип которого реализует этот метод:
```
fmt.Sprint arg e causes recursive call to (lab4/task4.ErrNegativeSqrt).Error method
```
`printf` входит в набор проверок, которые `go test` запускает автоматически, так что тесты с такой ошибкой даже не стартуют.

## Part 4: Generics and constraints

### Почему не компилируется
| # | Ошибка компилятора | Объяснение через type set | Исправление |
|---|---|---|---|
| 1 | `invalid operation: v == target (incomparable types in type set)` | в type set `any` есть слайсы, map, функции — для них `==` нет | `[T comparable]` |
| 2 | `invalid operation: a > b (type parameter T cannot use operator >)` | в `comparable` есть `bool`, структуры, указатели — у них только `==`/`!=` | `[T cmp.Ordered]` |
| 3 | `cannot use type Number outside a type constraint: interface contains type constraints` | интерфейс с union (`~int \| ~float64`) — только constraint, не обычный тип: значения такого типа не бывает | использовать в `[T Number]` или `var n int = 5` |
| 4 | `in call to Zero, cannot infer T` | `T` есть только в результате, вывести его не из чего (тип `x` выводится из вызова, а не наоборот) | `x := Zero[int]()` |

### Task 5: Generic toolbox — [task5/main.go](task5/main.go)
```
Sum ints:       6
Sum floats:     4
Sum scores:     387
Sum type:       main.Score
77.4 true
1.5 true
0 false
9 true
rust true
0 false
90 true
2
-1
1
passed: [90 75 88]
```
- `Sum(scores)` возвращает `Score`, а не `int`: `T` выведен как `Score`.
- `Average` возвращает `float64`: среднее `[]int{1, 2}` — `1.5`, в `T` оно бы обрезалось. `float64(Sum(nums))` разрешено, потому что **каждый** тип из type set `Number` конвертируется в `float64`.
- `MaxOf`/`Average` для пустого слайса возвращают нулевое значение и `false` (comma-ok, как у map).
- **Без `~`:**
  ```
  Score does not satisfy Number (possibly missing ~ for int in Number)
  ```
  `int | int64 | float64` содержит ровно три типа. `Score` — другой тип, хотя underlying у него `int`. `~int` = «`int` и все типы с underlying `int`».

### Task 6: Benchmark — [task6/](task6/) (Apple M4, go1.27.1, медиана из 3 прогонов)
```
BenchmarkSumInts-10              5116899      234.9 ns/op        0 B/op      0 allocs/op
BenchmarkSumGeneric-10           5106038      236.0 ns/op        0 B/op      0 allocs/op
BenchmarkSumAny-10               4062088      295.9 ns/op        0 B/op      0 allocs/op
BenchmarkSumAnyWithBoxing-10      267538     4483   ns/op    22336 B/op    745 allocs/op
```

| Бенчмарк | ns/op | B/op | allocs/op |
|---|---|---|---|
| `SumInts` (concrete) | 235 | 0 | 0 |
| `Sum[int]` (generic) | 236 | 0 | 0 |
| `SumAny` | 296 | 0 | 0 |
| `SumAny(ToAny(nums))` | 4483 | 22336 | 745 |

**Почему generic так же быстр, как concrete?** Go компилирует дженерики через **GC shape stenciling**: для каждой «формы» типа (по underlying-типу; все указатели — одна форма) генерируется отдельная копия машинного кода. Для `int` компилятор сделал `Sum[go.shape.int]` (видно в `go test -gcflags=-m`: `can inline Sum[go.shape.int]`). Там `+=` — обычное целочисленное сложение, ровно как в `SumInts`, и функция так же встраивается в цикл бенчмарка. `Score` и `int` имеют одну форму, значит и один и тот же код. (Дженерики бывают медленнее, когда методы constraint-а вызываются через словарь (dictionary) или когда `T` — указатель/интерфейс.)

**Почему `SumAny` медленнее даже с 0 аллокаций?** Аллокации были один раз, до `b.ResetTimer()`, а чтение из интерфейса ничего не аллоцирует. Время уходит на другое:
- элемент `[]any` занимает 16 байт (указатель на тип + указатель на данные) против 8 у `[]int`, поэтому памяти читается вдвое больше;
- `n.(int)` на каждом элементе — сравнение указателя на тип с дескриптором `int` плюс ветка на панику;
- само число лежит не в слайсе, а по указателю: лишняя загрузка из памяти (indirection);
- такой цикл компилятор не может развернуть/векторизовать так же, как сложение плотного `[]int`.

Разница всего ~25%, потому что упакованные числа аллоцировались подряд и лежат в памяти рядом (кэш работает хорошо). В реальном коде, где значения разбросаны по куче, разница больше.

**Откуда 745 allocs/op, а не 1001?** Числа в слайсе — `0, 1, …, 999`.
- 1 аллокация — `make([]any, 1000)`: 16000 байт, size class 16384.
- Чтобы положить `int` в интерфейс, нужен указатель на значение, поэтому `runtime.convT64` выделяет 8 байт в куче (`-gcflags=-m`: `n escapes to heap`).
- **Но:** для значений `< 256` `convT64` не аллоцирует, а отдаёт указатель на статический массив `runtime.staticuint64s` (а для нуля любого типа — на `runtime.zeroVal`). Числа `0…255` — 256 штук — бесплатны.
- Итого: 1 + (1000 − 256) = **745**. Байты сходятся: 16384 + 744 × 8 = **22336**.

Проверка: со слайсом `1000…1999` получается ровно **1001** allocs/op (24384 B/op), а со слайсом из одних нулей — **1** alloc/op.

Главная цена `any` — упаковка (boxing): 4483 ns против 296 ns, примерно в 15 раз медленнее.

### Task 7: Generic types — [task7/main.go](task7/main.go)
```
[] 0
[10 20 30] 3
[go is fun] 3
```
- `Push` добавляет в конец и **возвращает голову**, как `append`: `l = l.Push(v)`. Метод с pointer receiver можно вызвать на nil-указателе (как `Square.Area` в Part 3), но изменить переменную вызывающего он не может, поэтому для пустого списка новый узел возвращается.
- `Len` и `String` тоже корректно работают на `nil`: цикл `for ; l != nil; l = l.next` просто не выполняется.
- `fmt.Println(nums)` при `nums == nil` печатает `[]`: в `any` лежит `(*List[int], nil)` — это **не** nil-интерфейс, поэтому `fmt` вызывает наш `String()`.
- Методы используют параметр типа (`List[T]`). Своих параметров типа у метода по методичке быть не может — это верно для Go 1.18–1.26. В Go 1.27 generic-методы появились (нужно `go 1.27` в `go.mod`), но они не могут реализовывать методы интерфейсов. Подробнее — [THEORY.md, раздел 13.2](THEORY.md#132-могут-ли-методы-иметь-собственные-параметры-типа).
- Компромисс: `Push` в конец — O(n) (нет указателя на хвост). Push в начало был бы O(1), но печатал бы список в обратном порядке.

## ⭐ Бонус
```
true
false
panic: runtime error: comparing uncomparable type []int
```
- `a == b`: динамические типы одинаковые (`int`), значения равны, поэтому `true`.
- `c == d`: типы разные (`int` и `int64`), поэтому `false`, значения даже не сравниваются.
- `e == f`: типы одинаковые (`[]int`), Go пытается сравнить значения, а у слайсов нет `==`, поэтому **паника в рантайме**.

**Почему компилируется?** Интерфейсные типы сравнимы на уровне компилятора: `==` для `any` разрешён всегда. Что лежит внутри, известно только в рантайме, поэтому проверка переезжает туда. Та же ловушка: `map[any]int` с ключом-слайсом, а с Go 1.20 `any` удовлетворяет `comparable`, так что `Contains[any]` компилируется и тоже может паниковать.

## Теоретические вопросы к защите

**Что такое интерфейс? Как тип его реализует? Чем отличается от Java?**
Интерфейс — набор сигнатур методов (поведение, не данные). Тип реализует интерфейс **неявно**: достаточно, чтобы в его method set были все методы. В Java класс обязан написать `implements` при объявлении, то есть автор класса должен знать про интерфейс заранее. В Go интерфейс обычно объявляет **потребитель** и может описать уже существующие чужие типы (`*os.File` реализует твой `io.Writer`-подобный интерфейс, ничего о нём не зная). Поэтому интерфейсы в Go маленькие: 1–2 метода.

**Какие две вещи содержит значение интерфейса?**
Динамический тип и значение. Внутри это два машинных слова: указатель на itab (тип + таблица методов; у `any` — просто указатель на тип) и указатель на данные.

**`v := i.(T)` vs `v, ok := i.(T)`?**
Первая форма при неверном типе паникует (`interface conversion`). Вторая не паникует: возвращает нулевое значение `T` и `ok == false`.

**Какой тип у `x` в `case int, string:`? Почему?**
Тип выражения в switch (`any`). Переменная имеет один статический тип, а компилятор не может выбрать `int` или `string`. В case с одним типом `x` имеет именно этот тип.

**Когда интерфейс nil? Почему функция может вернуть не-nil error с nil-указателем?**
Только когда и тип, и значение — nil. При `return err`, где `err` типа `*ValidationError`, Go конвертирует указатель в интерфейс и записывает в него тип `*ValidationError`. Получается пара `(*ValidationError, nil)`, а она `!= nil`. Правило: возвращать `error` и литерал `nil`.

**Что такое type parameter и constraint?**
Type parameter (`T` в `[T Number]`) — параметр-тип, вместо которого при вызове подставляется конкретный тип (явно `Sum[int64]` или через вывод). Constraint — интерфейс, который задаёт **множество допустимых типов** и тем самым, какие операции разрешены над `T` внутри функции.

**`any` vs `comparable` vs `cmp.Ordered`?**
- `any` — все типы; можно только присваивать, передавать, возвращать.
- `comparable` — типы с `==`/`!=` (без слайсов, map, функций); можно использовать как ключ map.
- `cmp.Ordered` — целые, вещественные, строки (и их `~`); разрешены `<`, `<=`, `>`, `>=`.

**Что значит `~` в constraint?**
`~float64` — «`float64` и все типы, у которых underlying-тип `float64`» (`Celsius`, `Score`, …). Без `~` подходит только сам `float64`.

**Когда интерфейс, а когда дженерики?**
Интерфейс — когда у типов **разное поведение** (`Rect.Area` ≠ `Circle.Area`), нужна смешанная коллекция (`[]Shape`) или нужны только методы (`io.Writer`, `error`). Дженерики — когда **код одинаковый**, отличается только тип: контейнеры (`Stack[T]`, `List[T]`), алгоритмы (`slices.Contains`, `Sum`), когда нужны операторы `+`, `<`, `==`, или когда хочется избежать boxing в `any` (Task 6). Правило из Go blog: сначала пиши с конкретными типами, а дженерик вводи, когда поймал себя на копировании кода для другого типа.
