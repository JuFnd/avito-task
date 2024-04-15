### Описание архитектуры:
   Реализована микросевисная архитектура, общение сервисов по gRPC.
   - Реализовано два микросервиса:
   - Авторизация:
        Авторизация реализована на основе сессий.
        ![UNXyORX2pQ8](https://github.com/JuFnd/avito-task/assets/109366718/0a8f1eaa-9af5-4eef-bfc2-df2969b1bc46)

        - Схема БД:

          ![изображение](https://github.com/JuFnd/avito-task/assets/109366718/a36e0419-5f02-4d8d-a069-87d5304ffafd)

        - СУБД: Postgresql
        - БД Кэширования: Redis
   - Баннеры:
        - Схема БД:
     
        ![изображение](https://github.com/JuFnd/avito-task/assets/109366718/985e4b4e-4858-44f8-932c-0399120a5773)

        - СУБД: Postgresql

### Запросы
localhost:8081/api/v1/user_banner?tag_id=1&feature_id=1 GET
localhost:8081/api/v1/user_banner?tag_id=3&feature_id=3 GET
localhost:8081/api/v1/user_banner?tag_id=1&feature_id=1&use_last_revision=true GET

localhost:8081/api/v1/banner?feature_id=2&tag_id=2&limit=10&offset=0 GET

localhost:8081/api/v1/banner?feature_id=1&tag_id=2&limit=10&offset=0 GET



localhost:8081/api/v1/banner POST
Body:
```
{
   "tag_ids": [
      2
   ],
   "feature_id": 4,
   "content": "{\"content\": \"Banner 3 - Version 1\"}"
}
```

localhost:8081/api/v1/banner/12 PATCH
```
Body:
{
   "tag_ids": [
      1
   ],
   "feature_id": 4,
   "content": "{\"content\": \"Banner 4  - Version 1\"}"
}
```


localhost:8081/api/v1/banner/12 DELETE
```
Body:
{
   "tag_ids": [
      1
   ],
   "feature_id": 4,
   "content": "{\"content\": \"Banner 4  - Version 1\"}"
}
```


localhost:8080/signup
localhost:8080/signin
localhost:8080/logout
```
{
   "login": "test",
   "password": "test"
}
```
