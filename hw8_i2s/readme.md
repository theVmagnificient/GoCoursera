Функция interfaсe2struct

i2s - interface to struct. Функция, которая заполняет значения структуры из map[string]interface{} и подобных - того что получается, если распаковать json в interface{} (см. пример в json/dynamic.go)

Задание на работу с рефлексией.

Не смотря на некую мудрённость на первый взгляд - рефлексия применяется очень часто. Понимать как она работает и как вам с ней работать очень пригодится в дальнейшем.

Реализация занимает 80-100 строк кода

Из типов данных достаточно предусмотреть те, что есть в тесте.

Запускать go test -v

Код писать в файле i2s.go

Подсказки:

* Все нужные вам функции есть в пакете reflect - https://golang.org/pkg/reflect/ - внимательно читайте документацию
* json распаковывает int во float. Это указано в документации, не бага. В данном случае будет корректно приводить к инту, если нам встретился флоат
* Проверяйте всегда что вам приходит на вход. И смотрите, что вы передаёте в функцию (да, рекурсия тут себя хорошо показывает) не reflect.Value, а именно оригинальные данные, до который вы доковырялись через нужные методы рефлекта
* Если вы в функции используете какие-то имена структур, которые встречаются в стесте - это не правильно

```
=== RUN   TestSimple
struct
Username
string
Active
bool
ID
int
--- PASS: TestSimple (0.00s)
=== RUN   TestComplex
struct
SubSimple
struct
true
struct
Active
bool
ID
int
Username
string
AAAA
ManySimple
slice
true
slice
slice
main.Simple
*reflect.rtype
true
struct
ID
int
Username
string
Active
bool
55
&{42 rvasily true}
<main.Simple Value>
1
slice
main.Simple
*reflect.rtype
true
struct
ID
int
Username
string
Active
bool
55
&{42 rvasily true}
<main.Simple Value>
2
AAAA
Blocks
slice
true
slice
slice
main.IDBlock
*reflect.rtype
true
struct
ID
int
55
&{42}
<main.IDBlock Value>
1
slice
main.IDBlock
*reflect.rtype
true
struct
ID
int
55
&{42}
<main.IDBlock Value>
2
AAAA
--- PASS: TestComplex (0.00s)
=== RUN   TestSlice
slice
slice
main.Simple
*reflect.rtype
true
struct
Username
string
Active
bool
ID
int
55
&{42 rvasily true}
<main.Simple Value>
1
slice
main.Simple
*reflect.rtype
true
struct
Active
bool
ID
int
Username
string
55
&{42 rvasily true}
<main.Simple Value>
2
--- PASS: TestSlice (0.00s)
=== RUN   TestErrors
struct
ID
int
Username
string
Active
bool
struct
Username
string
Active
bool
ID
int
struct
Username
string
struct
SubSimple
struct
true
struct
ID
int
Username
string
Active
bool
AAAA
ManySimple
slice
true
slice
map
struct
SubSimple
struct
true
struct
bool
struct
slice
struct
--- PASS: TestErrors (0.00s)
PASS
```
