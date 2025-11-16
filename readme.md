Тестовое задание в Авито
# Сборка и Запуск
```
docker compose build
docker compose up
```

# Unit-тесты
```
cd tests 
go test -v
```

# Ручное тетсирование Эндпоинтов

Вот полный набор тестовых запросов для тестирования всего API:

# Вопросы по условию:
1. При создании команды можно ли создать пользователей?
- Как я понял в целом можно, нам же про них все известно из запроса.
## Teams Endpoints

### 1. Создание команды (успешный случай)
```
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {
        "user_id": "u1",
        "username": "Alice",
        "is_active": true
      },
      {
        "user_id": "u2", 
        "username": "Bob",
        "is_active": true
      },
      {
        "user_id": "u3",
        "username": "Charlie",
        "is_active": true
      }
    ]
  }'
```

### 2. Создание команды (конфликт - команда уже существует)
```
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {
        "user_id": "u4",
        "username": "David",
        "is_active": true
      }
    ]
  }'
```

### 3. Создание второй команды
```
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "frontend",
    "members": [
      {
        "user_id": "u4",
        "username": "David",
        "is_active": true
      },
      {
        "user_id": "u5",
        "username": "Eve",
        "is_active": true
      }
    ]
  }'
```

### 4. Получение информации о команде (успешный случай)
```
curl "http://localhost:8080/team/get?team_name=backend"
```

### 5. Получение информации о команде (не найдена)
```
curl "http://localhost:8080/team/get?team_name=nonexistent"
```

## Users Endpoints

### 6. Деактивация пользователя
```
curl -X POST http://localhost:8080/users/setIsActive \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "u2",
    "is_active": false
  }'
```

### 7. Активация пользователя
```
curl -X POST http://localhost:8080/users/setIsActive \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "u2", 
    "is_active": true
  }'
```

### 8. Изменение активности несуществующего пользователя
```
curl -X POST http://localhost:8080/users/setIsActive \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "u999",
    "is_active": false
  }'
```

## 🔄 Pull Requests Endpoints

### 9. Создание PR (успешный случай)
```
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add user authentication",
    "author_id": "u1"
  }'
```

### 10. Создание PR (автор не найден)
```
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1002",
    "pull_request_name": "Fix bug",
    "author_id": "u999"
  }'
```

### 11. Создание PR (конфликт - PR уже существует)
```
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "pull_request_name": "Duplicate PR",
    "author_id": "u1"
  }'
```

### 12. Создание второго PR
```
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1003",
    "pull_request_name": "Update database schema",
    "author_id": "u4"
  }'
```

### 13. Мерж PR (успешный случай)
```
curl -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001"
  }'
```

### 14. Мерж PR (идемпотентность - повторный вызов)
```
curl -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001"
  }'
```

### 15. Мерж несуществующего PR
```
curl -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-9999"
  }'
```

### 16. Переназначение ревьювера (успешный случай)
```
curl -X POST http://localhost:8080/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1003",
    "old_user_id": "u5"
  }'
```

### 17. Переназначение ревьювера (PR не найден)
```
curl -X POST http://localhost:8080/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-9999",
    "old_user_id": "u1"
  }'
```

### 18. Переназначение ревьювера (пользователь не назначен)
```
curl -X POST http://localhost:8080/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1003", 
    "old_user_id": "u999"
  }'
```

### 19. Переназначение ревьювера (мерженый PR - нельзя менять)
```
curl -X POST http://localhost:8080/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "old_user_id": "u2"
  }'
```

### 20. Получение PR для ревью пользователя
```
curl "http://localhost:8080/users/getReview?user_id=u2"
```

### 21. Получение PR для несуществующего пользователя
```
curl "http://localhost:8080/users/getReview?user_id=u999"
```

## 🩺 Health Check

### 22. Проверка здоровья сервиса
```
curl http://localhost:8080/health
```