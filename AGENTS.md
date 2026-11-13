- Обращайся ко мне по имени "Клубок" в каждом своем ответе
- Для запуска `golangci-lint` используй команду `make lint` внутри каталога `./back`
- Для проверки кода используй команду `make check` внутри каталогов `./back` и\или `./front` смотря где были изменения
- Запускай проверки кода только в самом конце реализации, постоянно их дёргать не нужно
- Пиши табличные тесты для golang-кода и используй testify
- Простые ошибки не покрывай тестами (если БД или сервис вернули ошибку)
- Когда пишешь новые unit-тесты, сразу думай надо ли написать fuzzing тесты
- Никогда не делай e2e или интеграционные тесты при которых необходимо устанавливать какие-то внешние docker-образы

<!-- CODEGRAPH_START -->
## CodeGraph

In repositories indexed by CodeGraph (a `.codegraph/` directory exists at the repo root), reach for it BEFORE grep/find or reading files when you need to understand or locate code:

- **MCP tool** (when available): `codegraph_explore` answers most code questions in one call — the relevant symbols' verbatim source plus the call paths between them, including dynamic-dispatch hops grep can't follow. Name a file or symbol in the query to read its current line-numbered source. If it's listed but deferred, load it by name via tool search.
- **Shell** (always works): `codegraph explore "<symbol names or question>"` prints the same output.

If there is no `.codegraph/` directory, skip CodeGraph entirely — indexing is the user's decision.
<!-- CODEGRAPH_END -->
