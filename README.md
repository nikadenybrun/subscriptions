## Subscription Service

Добрый день!

Данное приложение реализует систему для добавления и чтения постов и комментариев с использованием GraphQL на языке Golang.
---

## Описание проекта

Сервис предоставляет возможность:

-Можно просматривать список постов.<br>
-Можно просматривать пост с комментариями под ним.<br>
-Пользователь, написавший пост, может запретить оставление комментариев к своему посту.<br>
-Комментарии организованы иерархически, позволяя вложенность без ограничений.<br>
-Реализована система пагинации для получения списка комментариев.<br>
-Реализована возможность подписки на пост. Комментарии к постам доставляются асинхронно, т.е. клиенты, подписанные на определённый пост, получают уведомления о новых комментариях без необходимости повторного запроса.<br>
-Хранение данных реализовано в двух вариантах: в памяти (in-memory) и в PostgreSQL. Выбор варианта хранения должен быть реализован параметром при запуске приложения.<br>
---

## Технологии

- **Golang** - язык программирования
- **GraphQL** - язык программирования
- **PostgreSQL** - база данных для хранения информации о кошельках и балансах
- **Docker** - контейнеризация приложения и базы данных
- **docker-compose** - для поднятия системы

---

# Запуск

1. Клонируйте репозиторий.<br>

2. Запустите систему через docker-compose: docker-compose --env-file config.env up --build<br>

3. Для выбора варианта хранилища, нужно указать в config.env параметр {STORAGE_TYPE} как `in-memory` или `postgres', в зависимости от предпочтений<br>

4. Сервис будет доступен по адресу `http://localhost:8080`, база данных на порту `:5432`<br>

---

## Тестирование

Для запуска тестов используйте: 
    1. go test internal/tests/post_test.go -v<br>
    2. go test internal/tests/comment_test.go -v<br>

Для GraphQL playground по адресу `http://localhost:8080` можно использовать такие запросы:
1. Создание поста
 `mutation {
   createPost(title: "1 пост", content: "Это содержимое нового поста", commentsAllowed: true) {
     id
     title
     content
     commentsAllowed
   }
 }`
2. Создание обычного комментария
 `mutation {
   createComment(postID: "УНИКАЛЬНЫЙ ИДЕНТИФИКАТОР ПОСТА", text: "1 комментарий") {
     id
     postID
     text
     createdAt
   }
 }`
3. Создание комментария - ответа
 `mutation {
   createComment(postID: "УНИКАЛЬНЫЙ ИДЕНТИФИКАТОР ПОСТА", text: "1 вложенный комментарий", parentID: "УНИКАЛЬНЫЙ ИДЕНТИФИКАТОР ИЗНАЧАЛЬНОГО КОММЕНТАРИЯ") {
     id
     postID
     text
     createdAt
   }
 }`
4. Запрос поста с комментариями и пагинацией, где after - необязательный параметр
 `query {
   post(id: "УНИКАЛЬНЫЙ ИДЕНТИФИКАТОР ПОСТА", first: 3, after: "УНИКАЛЬНЫЙ ИДЕНТИФИКАТОР КОММЕНТАРИЯ, ПОСЛЕ КОТОРОГО БУДЕТ ВЫВЕДЕНЫ ОСТАЛЬНЫЕ") {
     id
     title
     content
     commentsAllowed
     comments(first: 3) {
       edges {
         cursor
         node {
           id
           text
           parentID
           createdAt
         }
       }
       endCursor
       hasNextPage
     }
   }
 }`
5. Запрос всех постов
 `query {
   posts {
     id
     title
     content
   }
 }`
6.Подписка на определенный пост
 `subscription {
   commentAdded(postID: "УНИКАЛЬНЫЙ ИДЕНТИФИКАТОР ПОСТА") {
     id
     postID
     text
     parentID
     createdAt
   }
 }`