# Lab 4 — теория: интерфейсы и дженерики

Подробный конспект к защите. Ответы на конкретные задания и вопросы из методички — в [ANSWERS.md](ANSWERS.md).
Все выводы программ и ошибки компилятора в этом файле проверены на go1.27.1 (Apple M4).

**Оглавление**

1. [Интерфейс: что это и как тип его реализует](#1-интерфейс-что-это-и-как-тип-его-реализует)
2. [Что лежит внутри значения интерфейса](#2-что-лежит-внутри-значения-интерфейса)
3. [Пустой интерфейс `any`](#3-пустой-интерфейс-any)
4. [Маленькие интерфейсы и принципы проектирования](#4-маленькие-интерфейсы-и-принципы-проектирования)
5. [Встраивание интерфейсов](#5-встраивание-интерфейсов)
6. [Type assertion](#6-type-assertion)
7. [Type switch](#7-type-switch)
8. [nil-интерфейсы](#8-nil-интерфейсы)
9. [`error` — тоже интерфейс](#9-error--тоже-интерфейс)
10. [Дженерики: зачем они нужны](#10-дженерики-зачем-они-нужны)
11. [Параметры типа и вывод типов](#11-параметры-типа-и-вывод-типов)
12. [Constraints = множества типов](#12-constraints--множества-типов)
13. [Generic-типы и методы](#13-generic-типы-и-методы)
14. [Как Go компилирует дженерики](#14-как-go-компилирует-дженерики)
15. [Дженерики в стандартной библиотеке](#15-дженерики-в-стандартной-библиотеке)
16. [Интерфейс или дженерик?](#16-интерфейс-или-дженерик)
17. [Бенчмарки: что нужно понимать для Task 6](#17-бенчмарки-что-нужно-понимать-для-task-6)
18. [Вопросы к защите — развёрнутые ответы](#18-вопросы-к-защите--развёрнутые-ответы)
19. [Шпаргалка](#19-шпаргалка)

---

## 1. Интерфейс: что это и как тип его реализует

### 1.1 Интерфейс описывает поведение, а не данные

Интерфейс — это тип, который задаётся **списком сигнатур методов**. Полей в нём нет. Он отвечает на вопрос «что значение умеет делать», а не «из чего оно состоит».

```go
type Shape interface {
	Area() float64
	Perimeter() float64
}
```

Переменная типа `Shape` может хранить **любое** значение, у которого есть оба метода.

### 1.2 Неявная реализация

В Go нет ключевого слова `implements`. Тип реализует интерфейс автоматически, если в его **method set** есть все методы интерфейса с точно такими же сигнатурами.

```go
type Rect struct{ W, H float64 }

func (r Rect) Area() float64      { return r.W * r.H }
func (r Rect) Perimeter() float64 { return 2 * (r.W + r.H) }

var s Shape = Rect{3, 4} // Rect нигде не пишет, что он Shape
```

Это называют **structural typing**: совместимость определяется структурой (набором методов), а не объявленным родством. Похоже на «утиную типизацию» Python, но проверка происходит **на этапе компиляции**.

**Отличие от Java:**

| | Java | Go |
|---|---|---|
| Связь типа с интерфейсом | явная: `class Rect implements Shape` | неявная: достаточно иметь методы |
| Кто объявляет интерфейс | обычно автор класса / библиотеки | обычно **потребитель** — там, где интерфейс нужен |
| Можно ли «подогнать» чужой тип | нет, нужен адаптер | да: чужой тип уже реализует твой интерфейс, если методы совпали |
| Иерархии | глубокие, `extends` + `implements` | плоские, интерфейсы из 1–2 методов |

Пример силы неявной реализации: ты пишешь в своём пакете

```go
type ByteSink interface{ Write(p []byte) (int, error) }
```

и `*os.File`, `*bytes.Buffer`, `*strings.Builder`, `net.Conn` **сразу** его реализуют, хотя их авторы о твоём интерфейсе не знали.

### 1.3 Method sets и интерфейсы (повтор Week 4)

| Тип | Method set |
|---|---|
| `T` | методы с ресивером `T` |
| `*T` | методы с ресивером `T` **и** `*T` |

```go
type Circle struct{ R float64 }

func (c *Circle) Area() float64      { return math.Pi * c.R * c.R }
func (c *Circle) Perimeter() float64 { return 2 * math.Pi * c.R }

var _ Shape = &Circle{1} // ok
var _ Shape = Circle{1}  // ошибка: method Area has pointer receiver
```

**Почему так?** Значение, положенное в интерфейс, — это **копия**, и она **не адресуема**. Чтобы вызвать pointer-метод, Go должен взять адрес `&c`, а у скрытой копии внутри интерфейса адреса нет. Даже если бы был, pointer-метод менял бы копию, а не исходную переменную — это была бы тихая ошибка. Поэтому компилятор запрещает это заранее.

### 1.4 Проверка на этапе компиляции

```go
var (
	_ Shape = Rect{}
	_ Shape = (*Circle)(nil)
)
```

- `_` — значение никуда не сохраняется, строка нужна только ради проверки типов.
- `(*Circle)(nil)` — nil-указатель **типа** `*Circle`: ничего не аллоцирует, но у него есть тип, и компилятор проверяет его method set.
- Если кто-то переименует метод или поменяет ресивер, сборка упадёт **здесь**, с понятным сообщением, а не где-то далеко в месте использования.

### 1.5 Полиморфизм и динамическая диспетчеризация

```go
func printInfo(s Shape) {
	fmt.Printf("%-14T area=%6.2f perimeter=%6.2f\n", s, s.Area(), s.Perimeter())
}

shapes := []Shape{Rect{3, 4}, &Circle{1}}
for _, s := range shapes {
	printInfo(s)
}
```
```
main.Rect      area= 12.00 perimeter= 14.00
*main.Circle   area=  3.14 perimeter=  6.28
```

`printInfo` компилируется **один раз** и не знает конкретных типов. Какой именно `Area` вызвать, решается **во время выполнения** по типу, лежащему внутри `s`. Это и есть полиморфизм в Go — без наследования, только через интерфейсы.

### 1.6 Value semantics vs pointer semantics при сохранении в интерфейс

Присваивание в интерфейс **копирует** то, что присваивается:

```go
type notifier interface{ notify() }
type user struct{ name string }

func (u user) notify() { fmt.Println("notify", u.name) }

u := user{"Bill"}
var n notifier = u   // в интерфейс скопировано ЗНАЧЕНИЕ
u.name = "Lisa"
n.notify()           // notify Bill — интерфейс хранит свою копию

var p notifier = &u  // в интерфейс скопирован УКАЗАТЕЛЬ
u.name = "Jack"
p.notify()           // notify Jack — данные общие
```

- Кладёшь **значение** — интерфейс владеет своей копией (value semantics).
- Кладёшь **указатель** — интерфейс разделяет данные с оригиналом (pointer semantics).

Правило method sets логично вытекает отсюда: если ты сказал «работай с копией» (положил значение), Go не даст вызвать метод, который ожидает **разделять** данные (pointer receiver).

---

## 2. Что лежит внутри значения интерфейса

### 2.1 Два машинных слова

Значение интерфейса — это пара **(динамический тип, значение)**. В рантайме — два указателя (16 байт на 64-битной платформе):

```go
// Интерфейс с методами (Shape, error, io.Writer...)
type iface struct {
	tab  *itab          // тип + таблица методов
	data unsafe.Pointer // указатель на данные
}

// Пустой интерфейс (any): методов нет, таблица не нужна
type eface struct {
	_type *_type        // только дескриптор типа
	data  unsafe.Pointer
}

// itab (runtime / internal/abi.ITab), упрощённо
type itab struct {
	inter *interfacetype // какой интерфейс (Shape)
	_type *_type         // какой конкретный тип (*Circle)
	hash  uint32         // копия хеша типа — для быстрых type switch
	fun   [1]uintptr     // адреса методов; на самом деле массив переменной длины
}
```

Три состояния, которые нужно уметь нарисовать на защите:

```
var s Shape             s = Rect{1, 2}           var sq *Square; s = sq
┌──────┬──────┐         ┌───────────┬──────┐     ┌──────────────┬──────┐
│ tab  │ data │         │ tab       │ data │     │ tab          │ data │
│ nil  │ nil  │         │ itab Rect │  ●   │     │ itab *Square │ nil  │
└──────┴──────┘         └───────────┴──┼───┘     └──────────────┴──────┘
                                       ▼
                                     {1 2}
s == nil → true         s == nil → false         s == nil → false
%T → <nil>              %T → main.Rect           %T → *main.Square
s.Area() → panic        s.Area() → 2             s.Area() → 0 (nil-ресивер)
```

### 2.2 Как вызывается метод через интерфейс

`s.Area()` компилируется примерно в:

```
f := s.tab.fun[i]   // i — номер Area среди методов интерфейса (они отсортированы по имени)
f(s.data)           // data передаётся как ресивер
```

Это **косвенный вызов** (indirect call): адрес функции читается из памяти. Он чуть дороже прямого и, главное, обычно **мешает инлайнингу** — компилятор не знает, какая функция будет вызвана. (Если компилятор сам видит, какой конкретный тип лежит в интерфейсе, он может сделать **девиртуализацию** и вызвать метод напрямую.)

### 2.3 Откуда берётся itab

- **Конверсия известна при компиляции** (`var s Shape = Rect{}`): компилятор знает и тип, и интерфейс, поэтому itab генерируется **статически**, как данные в бинарнике. В рантайме ничего не ищется.
- **Конверсия в рантайме** (`v.(Shape)`, где `v` — `any`): рантайм проверяет, есть ли у динамического типа все методы интерфейса, и строит itab (`runtime.getitab`). Результат кладётся в глобальный кэш, следующие проверки той же пары (тип, интерфейс) — быстрые.

### 2.4 Где хранится само значение (boxing)

Поле `data` — **всегда указатель** (так с Go 1.4). Отсюда два случая:

1. **Значение «размером с указатель»** — указатели, map, каналы, функции, структуры/массивы из одного указателя. Кладётся **прямо** в `data`, без аллокации.
2. **Всё остальное** — `int`, `float64`, `string`, структуры... Значение **копируется** в отдельную память, а в `data` кладётся указатель на копию. Это называется **boxing (упаковка)**. Если интерфейс «убегает» (escape analysis), копия выделяется в **куче**.

Рантайм экономит, где может. Проверено (`testing.AllocsPerRun`, значение сохраняется в глобальную `any`):

| Что кладём в `any` | allocs | Почему |
|---|---|---|
| константа `1000` | 0 | компилятор кладёт константу в read-only данные бинарника |
| переменная `int` = 7 | 0 | значения `0…255` берутся из статического массива `runtime.staticuint64s` |
| переменная `int` = 1000 | **1** | нужна копия 8 байт в куче (`runtime.convT64`) |
| `""` | 0 | нулевые значения → указатель на общий `runtime.zeroVal` |
| `"hi"` | **1** | копия заголовка строки (16 байт) |
| `struct{}{}` | 0 | у значений нулевого размера нечего копировать |
| структура 32 байта | **1** | копия в куче |
| `*big`, `map` | 0 | уже указатель — кладётся прямо в `data` |
| `true` | 0 | однобайтовые значения тоже из `staticuint64s` |

Именно это объясняет **745 аллокаций** в Task 6: числа `0…255` упаковываются бесплатно, `256…999` — по аллокации, плюс 1 на сам слайс.

### 2.5 Статический и динамический тип

- **Статический тип** — тип переменной в исходном коде (`Shape`). Его знает только компилятор, он определяет, **какие методы можно вызвать**.
- **Динамический тип** — тип значения, которое лежит внутри **сейчас** (`main.Rect`, `*main.Circle`). Меняется при каждом присваивании.

```go
var s Shape
fmt.Printf("%v %T\n", s, s) // <nil> <nil>
s = Rect{1, 2}
fmt.Printf("%v %T\n", s, s) // {1 2} main.Rect
s = &Circle{2}
fmt.Printf("%v %T\n", s, s) // &{2} *main.Circle
```

`%T` всегда печатает **динамический** тип: `Printf` получает аргументы как `...any`, и в `any` переносится только пара (динамический тип, значение) — статический тип `Shape` до `fmt` не доходит.

### 2.6 Сравнение интерфейсов `==`

Два значения интерфейса равны, если:
1. одинаковы их **динамические типы**, и
2. равны **значения** (по правилам `==` этого типа).

```go
var a, b any = 1, 1
a == b                 // true

var c, d any = 1, int64(1)
c == d                 // false — типы разные, значения даже не сравниваются

var e, f any = []int{1}, []int{1}
e == f                 // panic: runtime error: comparing uncomparable type []int
```

Компилятор разрешает `==` для **любых** интерфейсов: что внутри — он не знает. Поэтому для несравнимых типов (слайсы, map, функции) ошибка переезжает в **рантайм**. Та же ловушка с ключами map:

```go
m := map[any]int{}
m[[]int{1}] = 1 // panic: runtime error: hash of unhashable type []int
```

---

## 3. Пустой интерфейс `any`

`interface{}` — интерфейс без методов. Любой тип имеет «все ноль методов», значит **любой тип его реализует**. С Go 1.18 есть короткий псевдоним `any` (это то же самое, `type any = interface{}`).

```go
var x any
x = 42
x = "hello"
x = []int{1, 2}
```

**Цена `any`:**
- **Выключает проверку типов.** Положить можно что угодно, но чтобы что-то **сделать** со значением, нужно достать конкретный тип обратно (assertion / type switch), и ошибка вылезет только в рантайме.
- **Boxing**: не-указательные значения упаковываются, часто с аллокацией (см. 2.4).
- **Читаемость**: сигнатура `func Process(v any)` ничего не говорит о том, что функция ожидает. Go proverb: *«interface{} says nothing»*.

**Когда `any` уместен:** когда тип действительно неизвестен заранее — `fmt.Println(a ...any)`, `json.Marshal(v any)`, контейнеры для разнородных данных. Если же код одинаковый для разных, но **конкретных** типов, — это работа для дженериков (раздел 10+).

---

## 4. Маленькие интерфейсы и принципы проектирования

### 4.1 Главные интерфейсы стандартной библиотеки

| Интерфейс | Метод | Кто реализует |
|---|---|---|
| `fmt.Stringer` | `String() string` | всё, что хочет красиво печататься |
| `error` | `Error() string` | любое значение ошибки |
| `io.Writer` | `Write(p []byte) (n int, err error)` | `os.Stdout`, файлы, `*strings.Builder`, `*bytes.Buffer`, HTTP-ответы, `net.Conn` |
| `io.Reader` | `Read(p []byte) (n int, err error)` | `os.Stdin`, файлы, сетевые соединения, `strings.Reader`, тело HTTP-запроса |
| `io.Closer` | `Close() error` | файлы, соединения |
| `sort.Interface` | `Len`, `Less`, `Swap` | всё, что можно отсортировать через `sort.Sort` |

**Контракт `io.Reader`** (важная деталь): `Read` заполняет `p` и возвращает, сколько байт прочитано. Он может вернуть `n > 0` **и** ошибку одновременно (например, последние байты + `io.EOF`), поэтому сначала обрабатывают `n` байт, потом смотрят на `err`. `io.EOF` — не авария, а сигнал «данные закончились».

**Контракт `io.Writer`**: если `n < len(p)`, `Write` обязан вернуть ненулевую ошибку.

### 4.2 Сила маленького интерфейса

```go
func writeReport(w io.Writer, name string, score int) {
	fmt.Fprintf(w, "%s: %d points\n", name, score)
}

writeReport(os.Stdout, "Aruzhan", 95) // в терминал

var sb strings.Builder
writeReport(&sb, "Dias", 80)          // в строку
```

Одна функция — разные назначения: терминал, строка, файл, сеть, сжатый поток (`gzip.Writer`), хеш (`sha256.New()` тоже `io.Writer`), тестовый буфер. Функции ничего не нужно знать о получателе, кроме одного метода `Write`. Чем меньше метод, тем больше типов его реализуют.

### 4.3 Go proverbs и правила

- **«The bigger the interface, the weaker the abstraction»** (Rob Pike). Интерфейс из 10 методов реализует один тип — это не абстракция, а дублирование.
- **«Accept interfaces, return structs»**. Функция принимает минимальный интерфейс (гибкость для вызывающего), а возвращает конкретный тип (вызывающий получает все методы и сам решит, к какому интерфейсу его привести). Исключение — `error`: его всегда возвращают как интерфейс (см. раздел 8).
- **Объявляй интерфейс там, где он используется** (на стороне потребителя), а не рядом с реализацией. Тогда интерфейс содержит ровно те методы, которые нужны этому коду.
- **Не создавай интерфейс заранее**, «на будущее», если у него одна реализация. Сначала конкретный тип; интерфейс появится, когда понадобится вторая реализация или подмена в тестах.
- Имена интерфейсов из одного метода — по методу + `-er`: `Reader`, `Writer`, `Stringer`, `Notifier`.

---

## 5. Встраивание интерфейсов

Интерфейсы, как и структуры (Week 4), собираются из меньших:

```go
type Logger interface{ Log(msg string) }
type Saver  interface{ Save(data string) error }

type LogSaver interface {
	Logger
	Saver
}
```

Method set `LogSaver` — объединение методов `Logger` и `Saver`. Тип реализует `LogSaver`, если у него есть **все** эти методы. Именно так в стандартной библиотеке устроены:

```go
type ReadWriter interface {
	Reader
	Writer
}
type ReadWriteCloser interface {
	Reader
	Writer
	Closer
}
```

С Go 1.14 встраиваемые интерфейсы могут **пересекаться** по методам с одинаковой сигнатурой:

```go
type RWC interface {
	io.ReadCloser  // Read + Close
	io.WriteCloser // Write + Close — Close повторяется, это разрешено
}
```

Значение «большого» интерфейса можно присвоить переменной «меньшего» без assertion: `var r io.Reader = rw` (где `rw` — `io.ReadWriter`) — компилятор видит, что все методы есть.

---

## 6. Type assertion

### 6.1 Две формы

`i.(T)` — «я утверждаю, что в интерфейсе `i` лежит значение типа `T`».

```go
var i any = "Go"

s := i.(string)     // "Go" — угадали
n, ok := i.(int)    // 0 false — "comma ok": не паникует
f := i.(float64)    // panic: interface conversion: interface {} is string, not float64
```

| Форма | Если тип совпал | Если не совпал |
|---|---|---|
| `v := i.(T)` | `v` — значение типа `T` | **паника** |
| `v, ok := i.(T)` | `v` — значение, `ok == true` | `v` — нулевое значение `T`, `ok == false` |

Это тот же идиом «comma ok», что у map (`v, ok := m[k]`) и каналов (`v, ok := <-ch`). Однозначная форма уместна только там, где ошибка типа — **баг программы**, который нужно увидеть сразу.

Assertion применим только к значению **интерфейсного** типа: `x.(int)` для `x int` — ошибка компиляции.

### 6.2 Assertion к другому интерфейсу

`T` может быть интерфейсом. Тогда вопрос звучит иначе: **«есть ли у значения внутри эти методы?»**

```go
type Celsius float64

func (c Celsius) String() string { return fmt.Sprintf("%.1f°C", float64(c)) }

var t any = Celsius(36.6)
if st, ok := t.(fmt.Stringer); ok {
	fmt.Println("Stringer:", st.String()) // Stringer: 36.6°C
}
```

Результат `st` — новое значение интерфейса `fmt.Stringer` с тем же динамическим типом (`Celsius`) и теми же данными. Проверка выполняется в рантайме по method set динамического типа (с кэшированием itab, см. 2.3).

Так стандартная библиотека реализует **опциональные возможности**: функция принимает маленький интерфейс, а внутри проверяет, не умеет ли значение больше. Например, `io.Copy` проверяет, не реализует ли источник `io.WriterTo`, а приёмник — `io.ReaderFrom`, и если да — использует более быстрый путь.

### 6.3 Как `fmt` решает, что печатать

`fmt` для каждого аргумента делает серию assertion-ов (`handleMethods`) в таком порядке:

1. `fmt.Formatter` (свой метод `Format`) — полный контроль;
2. для `%#v` — `fmt.GoStringer` (`GoString()`);
3. **`error`** → вызывает `Error()`;
4. **`fmt.Stringer`** → вызывает `String()`.

Поэтому для типа, у которого есть и `Error()`, и `String()`, `fmt.Println` печатает `Error()`:

```go
type T struct{}
func (T) Error() string    { return "Error()" }
func (T) String() string   { return "String()" }
func (T) GoString() string { return "GoString()" }

fmt.Printf("%v | %s | %#v\n", T{}, T{}, T{}) // Error() | Error() | GoString()
```

Ещё одна защита `fmt`: если метод `Error()`/`String()` **паникует**, `fmt` перехватывает панику. Если при этом аргумент — nil-указатель, печатается `<nil>` (это объясняет вывод `error: <nil>` в Task 3); иначе — текст вида `%!v(PANIC=String method: ...)`.

### 6.4 Для ошибок — `errors.Is` / `errors.As`, а не голый assertion

Ошибки часто **оборачивают**: `fmt.Errorf("register user: %w", err)`. Обычный assertion смотрит только на внешний слой:

```go
wrapped := fmt.Errorf("register user: %w", &ValidationError{Field: "age"})

_, ok := wrapped.(*ValidationError) // false — снаружи *fmt.wrapError

var ve *ValidationError
errors.As(wrapped, &ve)             // true, ve.Field == "age" — разворачивает цепочку %w
```

- `errors.Is(err, target)` — есть ли в цепочке **конкретное значение** (sentinel: `io.EOF`, `ErrTooLong`).
- `errors.As(err, &target)` — есть ли в цепочке значение **типа** `target`; если да, записывает его в `target`. Второй аргумент — **указатель** на переменную (иначе `go vet` выдаст ошибку проверки `errorsas`).

---

## 7. Type switch

### 7.1 Синтаксис

Когда возможных типов несколько, вместо цепочки assertion-ов используют type switch:

```go
func describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case int, int64:
		return fmt.Sprintf("integer %v", x)          // x — any
	case string:
		return fmt.Sprintf("string of %d bytes", len(x)) // x — string
	case []int:
		return fmt.Sprintf("slice with %d ints", len(x)) // x — []int
	case error:
		return "error: " + x.Error()                 // x — error
	case fmt.Stringer:
		return "Stringer: " + x.String()             // x — fmt.Stringer
	default:
		return fmt.Sprintf("unknown type %T", x)     // x — any
	}
}
```

### 7.2 Правила

- Конструкция `v.(type)` разрешена **только** внутри `switch`.
- **Тип `x` в каждом case:**
  - один тип в case → `x` имеет **этот** тип;
  - несколько типов (`case int, int64:`) → `x` имеет **тип выражения** `v` (здесь `any`). Переменная в Go имеет один статический тип, а компилятор не может выбрать между `int` и `int64`. Поэтому `x + 1` там не компилируется: `invalid operation: x + 1 (mismatched types any and untyped int)`;
  - `default` и `case nil` → тоже тип `v`.
- Case проверяются **сверху вниз**, срабатывает **первый** совпавший. Значит, порядок важен, когда в case есть интерфейсы: тип `Timeout` с методами `Error()` и `String()` попадёт в тот из `case error` / `case fmt.Stringer`, который выше. Case с конкретным типом **ниже** подходящего интерфейса недостижим, и компилятор об этом **не** предупреждает.
- `case nil` срабатывает только для **nil-интерфейса** (нет динамического типа). Интерфейс с nil-указателем внутри попадёт в case своего типа.
- Дубли запрещены: `duplicate case int in type switch`.
- `fallthrough` в type switch запрещён.
- Можно не объявлять переменную: `switch v.(type) { ... }` — если в ветках значение не нужно.

### 7.3 Type switch или assertions?

Type switch — одна конструкция, `x` уже нужного типа, компилятор ловит дубли. Цепочка `if x, ok := v.(T); ok` — по `if` на каждый тип, `case int, int64` превращается в два `if`. Type switch читается лучше.

---

## 8. nil-интерфейсы

### 8.1 Главное правило

> Значение интерфейса равно `nil` **только** если и тип, и значение — `nil`.
> Интерфейс, в котором лежит nil-**указатель**, **не** равен `nil`.

```go
type Shape interface{ Area() float64 }
type Square struct{ Side float64 }

func (s *Square) Area() float64 {
	if s == nil {
		return 0
	}
	return s.Side * s.Side
}

var s Shape
fmt.Println(s == nil) // true  — (nil, nil)

var sq *Square
s = sq
fmt.Println(s == nil) // false — (*Square, nil)
fmt.Println(s.Area()) // 0     — метод вызван с nil-ресивером

s = nil
fmt.Println(s.Area()) // panic: invalid memory address or nil pointer dereference
```

### 8.2 Почему можно вызвать метод на nil-указателе, но не на nil-интерфейсе

- В `(*Square, nil)` **тип известен** → в itab есть адрес `(*Square).Area` → метод вызывается, `nil` передаётся как ресивер. Ресивер — это просто первый параметр функции, и `nil` — допустимое значение указателя. Если метод сам проверяет `s == nil`, всё работает (так устроены, например, методы `*List` в Task 7).
- В `(nil, nil)` **типа нет** → таблицы методов нет → искать `Area` негде → паника.

### 8.3 Классический баг: nil error, который не nil

```go
type ValidationError struct{ Field string }

func (e *ValidationError) Error() string { return "invalid field: " + e.Field }

func validate(age int) error {
	var err *ValidationError        // типизированный nil-указатель
	if age < 0 {
		err = &ValidationError{Field: "age"}
	}
	return err                      // ← здесь *ValidationError конвертируется в error
}

err := validate(20)
if err != nil {                     // true!
	fmt.Println("error:", err)      // error: <nil>
}
```

**Что происходит:** `return err` превращает `*ValidationError(nil)` в интерфейс `error`. При конвертации в интерфейс записывается **тип** `*ValidationError`, значит получается `(*ValidationError, nil)` ≠ `nil`. Как объясняет Go FAQ: интерфейс nil, только если **ничего** в него не положили; а здесь положили — указатель, пусть и nil.

Тот же баг без ошибок (пример из книги Донована и Кернигана):

```go
func write(out io.Writer) {
	if out != nil {                  // защита от nil...
		out.Write([]byte("done\n"))  // ...не помогла: panic, Write разыменовывает nil *Buffer
	}
}

var buf *bytes.Buffer                // забыли создать
write(buf)                           // в out лежит (*bytes.Buffer, nil)
```

**Правила, которые исключают баг:**
1. Функция возвращает интерфейс `error`, а не конкретный тип (`*ValidationError`).
2. Если ошибки нет — пишем **литерал** `return nil`.
3. Не храним типизированную переменную-указатель «на потом», чтобы вернуть её как интерфейс.

```go
func validate(age int) error {
	if age < 0 {
		return &ValidationError{Field: "age"}
	}
	return nil
}
```

### 8.4 Как доказать, что внутри nil-указатель

```go
fmt.Println(err == nil)                         // false
fmt.Printf("%T\n", err)                         // *main.ValidationError
fmt.Println(err.(*ValidationError) == nil)      // true — указатель внутри nil
```

Можно проверять «nil внутри» через `reflect.ValueOf(x).IsNil()`, но это лечение симптома. Правильно — не создавать такие значения (правила выше).

---

## 9. `error` — тоже интерфейс

```go
type error interface {
	Error() string
}
```

Любой тип с методом `Error() string` — это ошибка. Виды ошибок:

```go
// 1. Sentinel — заранее созданное значение, сравнивается через errors.Is
var ErrTooLong = errors.New("message too long")

// 2. Свой тип — несёт данные, достаётся через errors.As
type ValidationError struct{ Field string }
func (e *ValidationError) Error() string { return "invalid field: " + e.Field }

// 3. Обёртка — добавляет контекст, сохраняя исходную ошибку
return fmt.Errorf("sms to %s: %w", phone, ErrTooLong)
```

- `%w` в `fmt.Errorf` оборачивает ошибку (она доступна через `errors.Unwrap`, `Is`, `As`). `%v` — просто вставляет текст, цепочка теряется.
- `errors.Join(err1, err2)` (Go 1.20) объединяет несколько ошибок в одну; `Is`/`As` проверяют их все.

### 9.1 Ловушка: `fmt.Sprint(e)` внутри `Error()`

```go
type ErrNegativeSqrt float64

func (e ErrNegativeSqrt) Error() string {
	return fmt.Sprint("cannot Sqrt negative number: ", e) // ОШИБКА
}
```

`e` реализует `error` → `fmt` вызывает `e.Error()` (раздел 6.3) → тот снова вызывает `fmt.Sprint(e)` → снова `e.Error()` → ... Бесконечная рекурсия, стек растёт до лимита:

```
runtime: goroutine stack exceeds 1000000000-byte limit
fatal error: stack overflow
```

**Исправление** — конвертировать в тип **без методов**: `fmt.Sprint("...", float64(e))`. Тот же приём нужен в `String()`.

**`go vet` ловит это** анализатором `printf`: он знает функции семейства `fmt`, видит, что внутри метода `Error()` (или `String()`) в форматирующую функцию передаётся сам ресивер, тип которого этот метод определяет:

```
fmt.Sprint arg e causes recursive call to (lab4/task4.ErrNegativeSqrt).Error method
```

`printf` входит в подмножество проверок vet, которые `go test` запускает **автоматически** (`atomic, bools, buildtag, directive, errorsas, ifaceassert, nilfunc, printf, stdversion, stringintconv, tests`), поэтому с такой ошибкой тесты даже не стартуют.

---

## 10. Дженерики: зачем они нужны

До Go 1.18 (март 2022) для «одинакового кода с разными типами» было четыре плохих варианта:

1. **Копировать функцию** под каждый тип: `SumInts`, `SumFloats`, `SumInt64s`... Ошибку исправляешь в одной копии и забываешь в другой.
2. **`interface{}` + assertions**: `func Sum(nums []interface{}) interface{}` — пропадает проверка типов, нужен boxing (Task 6: в 15 раз медленнее с упаковкой), ошибки только в рантайме.
3. **Рефлексия** (`reflect`) — медленно и сложно читать.
4. **Кодогенерация** (`go generate`) — отдельный инструмент, сгенерированный код в репозитории.

Дженерики (параметрический полиморфизм) решают это: код пишется **один раз**, тип становится **параметром**, а компилятор проверяет каждый вызов.

```go
type Number interface {
	~int | ~int64 | ~float64
}

func Sum[T Number](nums []T) T {
	var s T // нулевое значение T: 0, 0.0, ...
	for _, n := range nums {
		s += n
	}
	return s
}

Sum([]int{1, 2, 3})        // 6   — T выведен как int
Sum([]float64{1.5, 2.5})   // 4   — T выведен как float64
Sum[int64]([]int64{10, 20}) // 30 — T указан явно
Sum([]string{"a"})         // ошибка компиляции: string does not satisfy Number
```

---

## 11. Параметры типа и вывод типов

### 11.1 Синтаксис

```go
func Map[T, U any](items []T, f func(T) U) []U {
	out := make([]U, 0, len(items))
	for _, it := range items {
		out = append(out, f(it))
	}
	return out
}
```

- `[T, U any]` — **список параметров типа**. `T`, `U` — имена; `any` — **constraint** (ограничение).
- Параметры типа объявляются в квадратных скобках **перед** обычными параметрами.
- Внутри функции `T` ведёт себя как тип: можно объявлять переменные `var x T`, слайсы `[]T`, map `map[K]V` и т. д.

### 11.2 Инстанциация

Подстановка конкретного типа вместо параметра называется **инстанциацией** (instantiation). Происходит в два шага:
1. Каждый параметр типа заменяется на аргумент-тип: `Sum[int]` → функция `func([]int) int`.
2. Компилятор проверяет, что аргумент-тип **удовлетворяет** constraint.

Обобщённую функцию нельзя использовать без инстанциации: `f := Sum` — ошибка (`cannot use generic function Sum without instantiation`), а `f := Sum[int]` — можно. Исключение: если тип очевиден из контекста присваивания, вывод сработает (`var f func([]float64) float64 = Sum` — компилируется).

### 11.3 Вывод типов (type inference)

Обычно аргументы типа писать не нужно — компилятор выводит их:

- **Из аргументов функции**: `Sum([]int{1, 2})` → `T = int`.
- **Нетипизированные константы** подстраиваются под типизированные аргументы; если все аргументы — нетипизированные константы разных видов, берётся «самый широкий» вид: `Max(1, 2.5)` → `T = float64`, результат `2.5`.
- **Из constraint-а** (constraint type inference) — когда один параметр выражен через другой:

```go
// Плохо: возвращает []E, и именованный тип теряется
func Scale[E constraints.Integer](s []E, c E) []E { ... }

type Point []int32
func (p Point) String() string { ... }

r := Scale(Point{1, 2}, 2)
r.String() // ошибка: r.String undefined (type []int32 has no field or method String)

// Хорошо: S — сам тип слайса, E выводится из constraint ~[]E
func Scale[S ~[]E, E constraints.Integer](s S, c E) S { ... }
// теперь Scale(Point{1, 2}, 2) возвращает Point
```

Так устроены функции пакета `slices`: `func Sort[S ~[]E, E cmp.Ordered](x S)`.

**Чего вывод не умеет:** выводить тип **только** из результата.

```go
func Zero[T any]() T { var z T; return z }

x := Zero()          // in call to Zero, cannot infer T
var y int = Zero()   // тоже ошибка — тип результата не используется для вывода
x := Zero[int]()     // ok
```

### 11.4 Нулевое значение `T`

```go
var zero T    // самый читаемый способ
*new(T)       // однострочник, например в return
```

С `T` нельзя сравнивать `nil` или `0`, если constraint это не разрешает: литерал `0` подходит не к любому `T`, а `nil` — только к указателям/слайсам/map/...

### 11.5 Assertion и type switch внутри дженерика

Значение типа `T` — не интерфейс, поэтому assertion напрямую не работает:

```go
func F[T any](v T) int { return v.(int) }
// invalid operation: cannot use type assertion on type parameter value v
```

Нужно сначала положить его в интерфейс:

```go
func Kind[T any](v T) string {
	switch any(v).(type) {
	case int:
		return "int"
	case string:
		return "string"
	}
	return "other"
}
```

Но если внутри дженерика появляется type switch по `T` — это сигнал, что поведение для разных типов **разное**, и, скорее всего, нужен интерфейс, а не дженерик (раздел 16).

---

## 12. Constraints = множества типов

### 12.1 Constraint — это интерфейс

Constraint — это **интерфейс**. Но для дженериков интерфейс нужно понимать как **множество типов** (type set), а не как список методов:

- `interface{ String() string }` — множество всех типов, у которых есть `String() string`;
- `interface{ int | float64 }` — множество из ровно двух типов;
- `interface{ ~int | ~float64 }` — `int`, `float64` и все типы, построенные на них;
- `any` (`interface{}`) — множество **всех** типов.

Constraint определяет две вещи:
1. **Какие типы** можно подставить вместо `T`.
2. **Какие операции** разрешены над `T` внутри функции — только те, что поддерживают **все** типы множества.

### 12.2 Элементы constraint-а

```go
// Методы
type Stringer interface{ String() string }

// Union (объединение) конкретных типов
type IntOrFloat interface{ int | float64 }

// Union с тильдой
type Number interface{ ~int | ~int64 | ~float64 }

// Типы + методы вместе: пересечение множеств
type StringableInt interface {
	~int
	String() string
}
```

Последний пример — «типы с underlying `int`, у которых **есть** метод `String`». Внутри функции доступны и арифметика, и метод:

```go
type Score int
func (s Score) String() string { return fmt.Sprintf("%d pts", int(s)) }

func Show[T StringableInt](v T) string {
	return v.String() + fmt.Sprint(" / +1 = ", v+1)
}

Show(Score(5)) // 5 pts / +1 = 6 pts
```

**Inline-constraint:** если constraint нужен в одном месте, можно не объявлять отдельный тип. `[T interface{ ~int | ~string }]` сокращается до `[T ~int | ~string]`.

### 12.3 Таблица основных constraint-ов

| Constraint | Какие типы | Что можно делать с `T` |
|---|---|---|
| `any` | все | присвоить, передать, вернуть, положить в `any` — **и всё** |
| `comparable` | типы с `==` (без слайсов, map, функций) | `==`, `!=`, ключ map |
| `cmp.Ordered` (Go 1.21) | все целые, вещественные, строки (с `~`) | `<`, `<=`, `>`, `>=`, `==`, `!=` |
| `int \| float64` | ровно `int` и `float64` | операции, общие для всех: `+ - * /`, `<`, `==`... |
| `~int \| ~float64` | `int`, `float64` и все типы на их основе | то же самое |
| `fmt.Stringer` | типы с методом `String() string` | вызвать `.String()` |

### 12.4 Тильда `~`

`~T` означает «`T` **и все типы, у которых underlying-тип `T`**».

```go
type Celsius float64          // underlying type — float64
type Score int                // underlying type — int

type Number interface{ int | float64 }   // без тильды
Sum([]Celsius{20.5, 21.5})
// Celsius does not satisfy Number (possibly missing ~ for float64 in Number)
```

`Celsius` — **отдельный** тип, его нет в множестве `{int, float64}`. С `~float64` множество включает все типы с underlying `float64`, и `Celsius` подходит. Реальные программы полны таких типов (`Celsius`, `Score`, `UserID`, `time.Duration` — это `~int64`), поэтому в constraint-ах почти всегда нужна `~`.

После `~` может стоять только тип, который **сам является своим underlying-типом** (`int`, `[]E`, `struct{...}`), — `~Celsius` написать нельзя.

### 12.5 Операции = пересечение по всем типам

Внутри дженерика операция разрешена, только если она есть у **каждого** типа множества. Разбор 4 программ из методички:

```go
// 1. any: в множестве есть слайсы, map, функции — у них нет ==
func Contains[T any](items []T, target T) bool { ... v == target ... }
// invalid operation: v == target (incomparable types in type set)
// → [T comparable]

// 2. comparable: в множестве есть bool, структуры, указатели — у них нет >
func Max[T comparable](a, b T) T { if a > b { ... } }
// invalid operation: a > b (type parameter T cannot use operator >)
// → [T cmp.Ordered]

// 3. Интерфейс с union — только constraint, не тип переменной
type Number interface{ ~int | ~float64 }
var n Number = 5
// cannot use type Number outside a type constraint: interface contains type constraints

// 4. T есть только в результате — выводить не из чего
func Zero[T any]() T { var z T; return z }
x := Zero()
// in call to Zero, cannot infer T  → Zero[int]()
```

**Конверсии** разрешены, если конверсия допустима для **каждого** типа множества: `float64(Sum(nums))` в `Average` компилируется, потому что и `int`, и `int64`, и `float64` (и всё на их основе) конвертируются в `float64`.

### 12.6 Ограничения на constraint-ы

```go
// Интерфейс с type-элементами нельзя использовать как обычный тип (см. пример 3 выше)

// В union нельзя ставить интерфейсы с методами
type Bad interface{ fmt.Stringer | int }
// cannot use fmt.Stringer in union (fmt.Stringer contains methods)

// Термы union не должны пересекаться
type MyInt int
type Bad2 interface{ ~int | MyInt }
// overlapping terms main.MyInt and ~int
```

Почему интерфейс-union нельзя использовать как тип переменной? Значение такого «типа» могло бы быть и `int`, и `float64`, и Go пришлось бы хранить его как интерфейс и проверять операции в рантайме — то есть снова `any` с assertion-ами. Суммы типов (sum types) в Go пока нет.

### 12.7 `comparable` подробнее

- `comparable` — встроенный constraint, его нельзя записать через union.
- **Strictly comparable** типы (числа, строки, указатели, каналы, структуры и массивы из strictly comparable полей) сравниваются без паники.
- **Интерфейсы** сравнимы, но **не строго**: `==` может паниковать (раздел 2.6). С **Go 1.20** такие типы тоже **удовлетворяют** `comparable`, поэтому:

```go
func Contains[E comparable](items []E, t E) bool { ... v == t ... }

Contains([]any{1, "a"}, any("a"))              // true
Contains([]any{[]int{1}}, any([]int{1}))       // panic: comparing uncomparable type []int
```

### 12.8 Где взять готовые constraint-ы

- Встроенные: `any`, `comparable`.
- `cmp.Ordered` (стандартная библиотека, Go 1.21).
- `golang.org/x/exp/constraints`: `Signed`, `Unsigned`, `Integer`, `Float`, `Complex`, `Ordered` (последний устарел в пользу `cmp.Ordered`).

---

## 13. Generic-типы и методы

### 13.1 Обобщённые типы

```go
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

var s Stack[string]   // инстанциация типа
nums := &Stack[int]{}
```

- `Stack` сам по себе — не тип, а «шаблон». Типы — `Stack[string]`, `Stack[int]`, и это **разные** типы.
- Для типов вывод аргументов не работает: всегда пишем `Stack[int]{}`.
- Ресивер метода **повторяет** параметры типа: `func (s *Stack[T]) Push(v T)`. Имена могут отличаться (`func (s *Stack[E]) ...`), constraint повторять не нужно — он берётся из объявления типа.
- Тип может ссылаться сам на себя: `type List[T any] struct { next *List[T]; val T }` (Task 7).

### 13.2 Могут ли методы иметь собственные параметры типа?

**Методичка:** нет — `func (b Box[T]) Map[U any](f func(T) U) Box[U]` даёт ошибку, и нужно писать обычную функцию `func Map[T, U any](items []T, f func(T) U) []U`. Это верно для **Go 1.18–1.26**, и именно это нужно отвечать на защите.

**Новое в Go 1.27** (твоя версия go1.27.1): в спецификацию добавлены **generic methods** — метод может объявить свои параметры типа. Но при `go 1.23` в `go.mod` компилятор говорит:

```
generic method requires go1.27 or later (-lang was set to go1.23; check go.mod)
```

С `go 1.27` в `go.mod` работает:

```go
type Box[T any] struct{ v T }

func (b Box[T]) Map[U any](f func(T) U) Box[U] { return Box[U]{f(b.v)} }

b := Box[int]{21}
s := b.Map(func(x int) string { return fmt.Sprint(x * 2) }) // Box[string]{"42"}
```

Ограничение осталось ровно то, из-за которого их так долго не добавляли: **интерфейсы**.

```go
type I interface { M[U any](U) }
// interface method must have no type parameters

type Mapper interface{ Map(func(int) string) string }
type Box struct{ v int }
func (b Box) Map[U any](f func(int) U) U { return f(b.v) }
var _ Mapper = Box{}
// Box does not implement Mapper (wrong type for method Map)
//     have Map[U any](func(int) U) U
//     want Map(func(int) string) string
```

**Почему:** при вызове через интерфейс конкретный тип неизвестен до рантайма, а инстанциацию `Map[string]` компилятор должен сделать **при компиляции** — заранее неизвестно, какие пары (тип, `U`) понадобятся. Поэтому generic-методы есть, но в реализации интерфейсов не участвуют.

---

## 14. Как Go компилирует дженерики

Есть два классических подхода:

| Подход | Где | Плюсы | Минусы |
|---|---|---|---|
| **Мономорфизация** (stenciling): отдельная копия кода на каждый тип | C++ templates, Rust | максимальная скорость | раздувание бинарника, медленная компиляция |
| **Стирание типов + boxing**: один код для `Object` | Java | один экземпляр кода | boxing, приведения типов в рантайме |

Go использует гибрид — **GC shape stenciling + dictionaries**:

- **GC shape** («форма») — то, как тип выглядит для сборщика мусора и аллокатора: размер, выравнивание, где указатели. Все типы с одинаковым **underlying-типом** имеют одну форму (`int`, `Score`, `UserID`), а **все указатели** — одну общую форму.
- Компилятор генерирует **одну копию кода на форму**: `Sum[go.shape.int]`, `Sum[go.shape.float64]`. Её видно в `go test -gcflags=-m`: `can inline Sum[go.shape.int]`.
- Различия между типами одной формы передаются скрытым аргументом — **словарём** (dictionary): дескрипторы типов, itab-ы для методов constraint-а и т. п.

**Последствия для скорости:**
- Для чисел и других не-указательных типов (`Sum[int]`) форма своя, `+=` — обычная машинная инструкция, функция инлайнится. Поэтому в Task 6 generic = concrete (236 vs 235 ns/op).
- Медленнее бывает, когда **вызываются методы constraint-а** (`v.String()` при `[T fmt.Stringer]`) или `T` — **указатель/интерфейс**: все указатели делят одну форму, и метод приходится искать через словарь — это косвенный вызов, который мешает инлайнингу. В таких случаях дженерик может быть даже медленнее обычного интерфейса.
- Вывод: дженерики в Go — инструмент для **устранения дублирования кода**, а не «бесплатная оптимизация».

---

## 15. Дженерики в стандартной библиотеке

Многое писать самому не нужно:

```go
slices.Contains([]string{"go", "rust"}, "rust") // true
slices.Index(s, v)                                // индекс или -1
slices.Max([]int{4, 9, 2})                        // 9 (паника на пустом слайсе)
slices.Sort(nums)                                 // сортировка по месту
slices.SortFunc(users, func(a, b User) int { return cmp.Compare(a.Age, b.Age) })
slices.BinarySearch(sorted, v)
slices.Reverse(s); slices.Equal(a, b); slices.Compact(s)

maps.Keys(m)                       // итератор (Go 1.23), не слайс!
slices.Sorted(maps.Keys(m))        // отсортированный слайс ключей
slices.Collect(maps.Values(m))     // слайс значений

cmp.Compare(a, b)                  // -1 / 0 / +1
cmp.Or(name, "anonymous")          // первое ненулевое значение

min(3, 8, 1); max(3, 8, 1)         // встроенные функции (Go 1.21)
clear(m)                           // очистить map / обнулить слайс (Go 1.21)
```

Ещё примеры: `sync/atomic.Pointer[T]` (Go 1.19), `sync.OnceValue[T]` (Go 1.21), `iter.Seq[V]` / `iter.Seq2[K, V]` (Go 1.23, итераторы для `range`), `unique.Handle[T]` (Go 1.23).

---

## 16. Интерфейс или дженерик?

| Интерфейс, когда... | Дженерик, когда... |
|---|---|
| у типов **разное поведение** (`Rect.Area` ≠ `Circle.Area`) | **код одинаковый** для всех типов, отличается только тип |
| нужна **смешанная** коллекция (`[]Shape` с прямоугольниками и кругами) | пишешь **контейнер**: стек, очередь, множество, дерево, список |
| ты только **вызываешь методы** | нужны **операторы** `+`, `<`, `==` |
| примеры: `io.Writer`, `error`, `fmt.Stringer` | примеры: `slices.Contains`, `Sum`, `Stack[T]`, `List[T]` |

Рекомендации из Go blog («When To Use Generics», Ian Lance Taylor):

- **Пиши код, а не проектируй типы.** Начни с конкретных типов. Если ловишь себя на том, что **копируешь** функцию для другого типа — тогда вводи параметр типа.
- Дженерики хороши для функций, работающих со **слайсами, map, каналами** любого типа элементов (`slices`, `maps`), и для **структур данных общего назначения**.
- Если всё, что делает функция, — **вызывает метод** значения, используй интерфейс:

  ```go
  // Плохо: дженерик ради вызова одного метода
  func ReadSome[T io.Reader](r T) ([]byte, error)
  // Хорошо
  func ReadSome(r io.Reader) ([]byte, error)
  ```

- Если **реализация разная** для разных типов (type switch по `T` внутри дженерика), — это интерфейс.
- Если тип вообще неизвестен и нужно инспектировать структуру (`encoding/json`) — рефлексия.

**Тонкая разница: `[T fmt.Stringer]` vs `[]fmt.Stringer`.**

```go
func JoinAll[T fmt.Stringer](items []T, sep string) string // дженерик
func JoinIface(items []fmt.Stringer, sep string) string     // интерфейс

cs := []Celsius{20.5, 22}
JoinAll(cs, ", ")    // ok — T = Celsius
JoinIface(cs, ", ")  // ошибка: cannot use cs (variable of type []Celsius) as []fmt.Stringer value
```

`[]Celsius` и `[]fmt.Stringer` — разные типы с разной памятью (8 байт на элемент против 16: пара тип+данные), и Go не конвертирует слайсы автоматически. Для интерфейсной версии пришлось бы создать новый `[]fmt.Stringer` и упаковать каждый элемент. Дженерик принимает слайс как есть. Зато в `[]fmt.Stringer` можно сложить **разные** типы, а `[]T` всегда однороден.

---

## 17. Бенчмарки: что нужно понимать для Task 6

```go
func BenchmarkSumAny(b *testing.B) {
	anys := ToAny(numbers(1000)) // подготовка
	b.ResetTimer()               // не считать подготовку
	for i := 0; i < b.N; i++ {
		sink = SumAny(anys)      // sink — глобальная переменная
	}
}
```

- `b.N` подбирает сам `testing`: запускает с 1, 100, 10000... пока замер не займёт ~1 секунду. `ns/op` = общее время / `b.N`.
- `b.ResetTimer()` обнуляет таймер и счётчики аллокаций — подготовка не попадает в результат.
- `-benchmem` (или `b.ReportAllocs()`) добавляет `B/op` и `allocs/op`. Считаются **только аллокации в куче**; копирование на стеке не видно (Week 4).
- **`sink`** — результат сохраняется в глобальную переменную, иначе компилятор может увидеть, что результат не используется, и выбросить вычисление (dead code elimination) — бенчмарк покажет ~0 ns.
- **Инлайнинг**: маленькие функции встраиваются в цикл бенчмарка — и `SumInts`, и `Sum[int]` в Task 6 встроились одинаково.
- `-count=N` повторяет замер N раз; разброс показывает, насколько цифрам можно верить. Для сравнения версий есть `benchstat`.
- С Go 1.24 есть `for b.Loop() { ... }` — сам исключает подготовку из замера и не даёт компилятору выбросить результат.

Как читать результаты Task 6 — в [ANSWERS.md](ANSWERS.md), раздел «Task 6: Benchmark».

---

## 18. Вопросы к защите — развёрнутые ответы

**1. Что такое интерфейс? Как тип его реализует? Чем это отличается от Java?**
Интерфейс — тип, заданный набором сигнатур методов; описывает поведение, а не данные. Тип реализует интерфейс неявно: если в его method set есть все методы интерфейса. Нет `implements` — связь проверяет компилятор в момент присваивания. В Java класс обязан явно объявить `implements`, то есть автор класса должен знать про интерфейс заранее. В Go интерфейс обычно объявляет потребитель, и он может описать уже существующие типы, даже из чужих пакетов (`*os.File` реализует любой интерфейс с методом `Write(p []byte) (int, error)`). Отсюда стиль Go: маленькие интерфейсы, объявленные там, где они используются. Нюанс method sets: если метод на `*T`, интерфейс реализует только `*T`.

**2. Какие две вещи содержит значение интерфейса?**
Динамический тип и значение. В памяти — два слова: указатель на itab (для `any` — просто на дескриптор типа) и указатель на данные. Itab содержит тип и таблицу адресов методов — по ней вызывается `s.Area()`. Не-указательные значения копируются (упаковываются) и хранятся по указателю, часто с аллокацией в куче. `%T` печатает динамический тип.

**3. Чем отличаются `v := i.(T)` и `v, ok := i.(T)`?**
Первая форма при неверном типе паникует (`interface conversion: interface {} is string, not float64`). Вторая не паникует: `ok == false`, `v` — нулевое значение `T`. Если `T` — интерфейс, проверяется, есть ли у динамического типа все его методы. Для ошибок вместо assertion используют `errors.As` — он разворачивает обёртки `%w`.

**4. Какой тип у `x` в `case int, string:`? Почему?**
Тип выражения в switch (обычно `any`). Переменная имеет один статический тип, а `x` может быть и `int`, и `string` — компилятор не может выбрать. Поэтому `x + 1` там не компилируется. В case с одним типом `x` имеет этот тип. В `default` и `case nil` — тоже тип выражения.

**5. Когда значение интерфейса nil? Почему функция может вернуть не-nil error с nil-указателем?**
Только когда и тип, и значение — nil. При `return err`, где `err` — `*ValidationError(nil)`, указатель конвертируется в `error`, и в интерфейс записывается тип `*ValidationError`. Получается `(*ValidationError, nil)`, а это не nil. Методы на таком значении вызываются (тип известен), но разыменование nil внутри метода — паника; `fmt` её перехватывает и печатает `<nil>`. Правила: возвращать тип `error`, литерал `nil`, не хранить типизированный указатель «на потом».

**6. Что такое type parameter? Что такое constraint?**
Type parameter (`T` в `func Sum[T Number]`) — параметр-тип; при вызове вместо него подставляется конкретный тип (инстанциация), явно `Sum[int64](...)` или через вывод из аргументов. Constraint — интерфейс, который задаёт множество допустимых типов для `T` и тем самым определяет, какие операции разрешены над `T` внутри функции: только общие для всех типов множества. Компилятор проверяет каждый вызов: `Sum([]string{})` не компилируется.

**7. Чем отличаются `any`, `comparable` и `cmp.Ordered`?**
- `any` — все типы; можно только присваивать, передавать, возвращать.
- `comparable` — типы с `==`/`!=` (без слайсов, map, функций); можно использовать как ключ map. С Go 1.20 сюда входят и интерфейсы, у которых `==` может паниковать.
- `cmp.Ordered` — целые, вещественные числа и строки (с `~`), поддерживают `<`, `<=`, `>`, `>=`. Каждое следующее множество — подмножество предыдущего.

**8. Что значит `~` в constraint?**
`~T` — «`T` и все типы, у которых underlying-тип `T`». `~float64` включает `Celsius`, `~int` — `Score`, `~int64` — `time.Duration`. Без `~` в множестве только сам тип, и вызов с `Celsius` даёт `Celsius does not satisfy Number (possibly missing ~ for float64 in Number)`.

**9. Когда выбрать интерфейс, а когда дженерики?**
Разное поведение → интерфейс: `Rect.Area` и `Circle.Area` считаются по-разному, нужен `[]Shape` со смешанными типами, функция только вызывает методы (`io.Writer`). Одинаковый код, разные типы → дженерики: контейнеры (`Stack[T]`, `List[T]`), алгоритмы над слайсами и map (`slices.Contains`, `Sum`), нужны операторы `+`, `<`, `==`, хочется избежать boxing в `any` (Task 6: с упаковкой в 15 раз медленнее). Правило Go blog: сначала конкретные типы; дженерик — когда начал копировать код для другого типа.

**Возможные дополнительные вопросы:**

- *Зачем `var _ Shape = (*Circle)(nil)`?* — проверка при компиляции, без аллокации; ошибка видна в месте объявления.
- *Почему `Circle{}` не реализует `Shape`, если методы на `*Circle`?* — значение в интерфейсе — неадресуемая копия, Go не может подставить `&`.
- *Почему `fmt.Println(e)` для типа с `Error()` и `String()` печатает `Error()`?* — `fmt` проверяет `error` раньше `Stringer`.
- *Почему `fmt.Sprint(e)` в `Error()` — бесконечная рекурсия? Как ловит vet?* — раздел 9.1.
- *Почему generic `Sum` так же быстр, как `SumInts`?* — GC shape stenciling, раздел 14.
- *Почему 745, а не 1001 аллокация?* — `staticuint64s` для `0…255`, раздел 2.4.
- *Что напечатает `any([]int{1}) == any([]int{1})`?* — паника в рантайме, раздел 2.6.
- *Можно ли у метода свои параметры типа?* — до Go 1.26 нет; в Go 1.27 можно, но такой метод не реализует методы интерфейсов (раздел 13.2).

---

## 19. Шпаргалка

- Интерфейс — набор методов. Реализуется **неявно**, просто наличием методов.
- Значение интерфейса — пара **(динамический тип, значение)**. `%T` показывает динамический тип.
- Значение в интерфейсе — **копия**, неадресуемая: pointer-методы есть только у `*T`.
- Держи интерфейсы **маленькими**: у `io.Writer`, `error`, `fmt.Stringer` по одному методу.
- `i.(T)` паникует при ошибке, `v, ok := i.(T)` — нет. Для многих типов — **type switch**.
- В `case A, B:` переменная имеет тип выражения (`any`), в `case A:` — тип `A`. Порядок case важен.
- Интерфейс nil, только если **и тип, и значение** nil. Для «нет ошибки» — **литерал `nil`**.
- `fmt` проверяет `error` раньше `Stringer`; `fmt.Sprint(e)` внутри `Error()` — рекурсия, ловит `go vet`.
- Дженерики убирают копипасту: `func Sum[T Number](nums []T) T`.
- Constraint — **множество типов**; оно же решает, какие операторы разрешены над `T`.
- `~T` включает именованные типы вроде `type Celsius float64`.
- `any` ⊃ `comparable` ⊃ `cmp.Ordered`.
- Вывод типов работает по аргументам, **не по результату**: `Zero[int]()`.
- Go компилирует дженерики через **GC shape stenciling + dictionaries**: для чисел — так же быстро, как конкретный код.
- **Разное поведение → интерфейс. Одинаковый код, разные типы → дженерики.**
