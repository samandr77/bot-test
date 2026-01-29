# AI Telegram Bot (GPT, Sora, NanoBanana)

Современный и надежный Telegram-бот на языке Go, интегрирующий передовые модели искусственного интеллекта для генерации текста, видео и изображений.

## Ключевые возможности
- 🤖 **GPT-интеграция**: Интеллектуальные диалоги с использованием моделей OpenAI.
- 🎬 **Sora 2**: Генерация видео по текстовому описанию.
- 🍌 **NanoBanana**: Продвинутая генерация и редактирование изображений.
- 💰 **Баланс и Оплата**: Система подписки и пополнения баланса (интеграция с Yookassa).
- 🛠 **FSM**: Надежное управление состояниями пользователей через Redis.

## Технологический стек
- **Язык**: [Go 1.24+](https://go.dev/)
- **Библиотека бота**: [go-telegram/bot](https://github.com/go-telegram/bot) (современный подход с поддержкой Context и Middleware)
- **База данных**: PostgreSQL (хостинг на **Supabase**)
- **ORM**: [GORM](https://gorm.io/)
- **Миграции**: [goose](https://github.com/pressly/goose)
- **Кэширование/FSM**: [Redis Cloud](https://redis.io/)
- **Логирование**: Structured `slog` (JSON в prod, Text в dev)
- **Качество кода**: `golangci-lint`


## Быстрый старт

### 1. Настройка окружения
Создайте файл `.env` на основе примера:
```bash
cp env.example .env
```
Заполните обязательные параметры:
- `TELEGRAM_BOT_TOKEN`: Токен от @BotFather.
- `DATABASE_URL`: Строка подключения к PostgreSQL.
- `REDIS_URL`: URL для подключения к Redis.
- `OPENAI_API_KEY`: Ключ API OpenAI.
- `PAYMENT_PROVIDER_TOKEN`: Для приема платежей.

### 2. Запуск проекта
Используйте `Makefile` для работы с проектом:
```bash
make check  # Полная проверка (линт + тесты + сборка)
make run    # Запуск приложения
```

## Структура проекта
- `cmd/bot/`: Точка входа, инициализация зависимостей.
- `internal/bot/`: Обработчики команд Telegram, middleware.
- `internal/service/`: Бизнес-логика приложения.
- `internal/repository/`: Работа с БД и Redis.
- `internal/models/`: Структуры данных (GORM модели, UserState).
- `internal/config/`: Загрузка и валидация конфигурации.
