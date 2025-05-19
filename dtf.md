## Поиск по новостям

GET
https://api.dtf.ru/v2.8/search/posts?markdown=false&sorting=date&q=%D1%80%D0%BE%D0%B7%D1%8B%D0%B3%D1%80%D1%8B%D1%88&title=true&editorial=false&strict=false&dateFrom=1746539876

## Реакция

POST
https://api.dtf.ru/v2.5/content/{post_id}/react

BODY

```
type: 1
referer: discovery
```

## Комментарий

POST
https://api.dtf.ru/v2.4/comment/add

BODY

```
id: 2199774 // номер новости
text: Эх, не успел // комментарий
attachments: []
reply_to: 0
nonce: 416019 // ???
referer: discovery // ???
```

## Логин

POST
https://api.dtf.ru/v3.0/auth/email/login

BODY
05b482e46bb53425fcecfd139be42fa7a76d3a309f0d2b758ac901273c424de8

```
email: _____
password: _____
```

Notes:
Access Token живет 5 минут
Refresh Token живет 2 месяца

eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJqdGkiOiI2ODFiNmQwZjQwZjE0Mi43MzU2ODg1NSIsImlhdCI6MTc0NjYyNzg1NS4yNjU4ODYsImV4cCI6MTc0NjYyODE1NS4yNjU4ODYsInVzZXJfaWQiOjExMzU1Nywicm9sZSI6MCwibmFtZSI6IiIsImVtYWlsIjoiIiwiYmFubmVkIjpmYWxzZSwic3Vic2l0ZV9pZCI6MTEzNzMyfQ.1mMmMsmLYcrJ_kcN_XVa8m0kDDg3MWMBb_yNRUMaw_E
