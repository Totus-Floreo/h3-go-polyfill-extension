# h3-go-polyfill-extension

## Описание

Инструмент для генерации H3-полифиллов по геометрии из PostGIS. Позволяет брать мультиполигон из таблицы Postgres, разбивать его на ячейки H3 нужной формы, эксклюзивно входящих в фигуру с полным заполнением, и сохранять результат обратно в БД.

## Пример

[![Обычный режим работы h3.PolygonToCells](img/preview_normal.png)](img/normal.png)
[![Экспериментальный режим работы h3.PolygonToCells](img/preview_cob.png)](img/ContainmentOverlappingBbox.jpg)
[![Итог работы приложения](img/preview_modified.png)](img/modified.jpg)

## Возможности
- Получение мультиполигона из PostGIS
- Конвертация в H3-ячейки (через Uber H3)
- Сохранение результата в отдельную таблицу
- Гибкая настройка через YAML или флаги
- Логирование через slog

## Структура проекта
- `cmd/` — точка входа CLI
- `internal/app/` — бизнес-логика (Service)
- `internal/domain/` — интерфейсы сервисов и репозиториев
- `internal/repository/` — реализация работы с БД
- `internal/logging/` — логгер
- `internal/env/` — работа с переменными окружения и конфигом
- `pkg/util/` — утилиты для работы с H3 и геометрией

## Важно
- Работает только на UNIX-системах (Linux, macOS) из-за особенностей h3-go и cgo

## Требования
- Go 1.21+
- Postgres с расширением PostGIS

## Быстрый старт

1. Установите зависимости:
   ```sh
   go mod tidy
   ```
2. Подготовьте таблицы в Postgres (outline и целевую для H3)
3. Заполните `config.yaml` или используйте флаги:
   ```yaml
   target: h3
   outline: outline
   db_conn: "postgres://user:password@localhost:5432/db"
   ```
4. Запустите:
   ```sh
   go run main.go --config config.yaml
   # или с флагами
   go run main.go --target h3 --outline outline --conn "postgres://user:password@localhost:5432/db"
   ```

## Пример использования

```
go run main.go --config config.yaml
```

## Контакты
Автор: Timur Kulakov <totusfloreodev@proton.me>
