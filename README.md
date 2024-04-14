## Команды запуска в Makefile:
- `make docker` - запуск всего приложения и его зависимостей в докер-среде. Для запуска нужен докер
- `make run` - локальный запуск проекта. Для запуска нужен postgres, слушающий на 5432 порту, с настроенной политикой безопасности подключений. Скорее всего этот вариант Вам не подходит. Используется для локальной разработки
- `make docker-deps` - запускает все зависимости в докере, но само go-приложение запускается локально. Для этого нужна версия go >= 1.22.1, установленная локально
- `make lint` - запуск линтера для кода всего проекта. Для этого локально должен быть установлен golangci-lint. Конфигурация линтера описана в файле `.golangci.yml`
- `make test` - запускает тесты, предварительно подняв в докере все зависимости. Для запуска нужен докер. После того, как все тесты завершили свою работу, контейнер с зависимостями в докере автоматически останавливается и удаляется
- `jwt-user` - получение пользовательского jwt-токена
- `jwt-admin` - получение админского jwt-токена

## Комментарии по заданию
- В задании расплывчато было написано про токены пользователя и администратора. Насущные вопросы мелькают в голове: откуда их брать? нужно ли делать сервис авторизации/регистрации самому? Т.к. в описании интерфейса API нет и намёка на эндпоинты авторизации, я решил сделать обычное cmd-приложение, которое по запросу способно выдавать JWT-токены с нужной ролью: admin/user. Эти токены считаются валидными во всём приложении, и роль отправителя запроса определяется именно по такому токену, предоставленному в хэдере Authorization.

## Комментарии по решению
- Выполнены все обязательные пункты условия
- Из дополнительных заданий проведено нагрузочное тестирование (результаты лежат в tests/load). Не совсем понятно было, в каких условиях проводить это тестирование, какие эндпоинты нужно тестировать. Я произвёл два запуска: один на простой хэндлер, который возвращает 200 и пишет OK в ответ (скорее только лишь для определения верхней границы производительности приложения, запущенного на моей машине), и второй, более сложный, с сохранением сущности баннера. 216 запроса, которые отображаются как провалившиеся - это коллизии пары (фича, тег) с уже существующей в бд такой же записью (id для нагрузочного теста генерировались случайно, в бд предварительно было загружено необходимое количество данных фичей и тегов). Эти запросы не провалились, сервер просто вернул на них 400 Bad Request. 
- Также в папке tests лежат интеграционные тесты почти на все сценарии работы с приложением
- Конфигурация линтера `golangci-lint`, как уже написал, находится в файле `.golangci.yml`
- Коллекция с запросами Postman находится в файле `avito-task-banners-scilightener.postman_collection.json`
- В "Общих вводных" к заданию был такой пункт: "Фича и тег однозначно определяют баннер". Я счёл это как свойство самого домена, а не правило бизнес-логики, поэтому имплементировал его на уровне хранения данных - в самой бд посредством тригера (см. `migrations/1_init.up.sql`)

## Как бы улучшил существующее решение:
- Во-первых, добавил бы, конечно, кэширование, на что в задании прямым текстом намекалось. Я старался выделять интерфейсы по минимальному функционалу, поэтому добавить слой кэширования в сервис не составит большого труда: достаточно создать адаптер между реализацией репозитория и потенциального кэш-хранилища: в зависимости от параметра useLastRevision он будет передавать запрос в соответствующее место (репозиторий или кэш). Обновлять кэш несложно, достаточно для всех других методов создать такую же обёртку, которая будет перехватывать запрос на запись в бд, посылать его в отдельную несинхронизируемую го-рутину для обновления кэша, а сам запрос передавать дальше в репозиторий
- Реализацию удаления по фиче/тегу через отложенные операции я не успел завершить. Предполагалось, что запрос на удаление будет создавать команду с параметрами: featureID, tagID, которая будет посылаться в персистентную шину сообщений (например, RabbitMQ). И консьюмер на своей стороне уже в асинхронном режиме эту команду обрабатывать. Другая, более наивная и простая реализация была бы простым несинхронизированным запуском го-рутины с походом в базу и удалением данных, но такой подход чреват тем, что из-за возможных ошибок можно потерять контекст исполнения команды, и соответственно её не выполнить
- Версии баннеров я бы реализовал так: в таблицу баннеров с помощью новой миграции добавил столбец - текущая версия банера, и первичный ключ в этой таблице стал бы теперь составным - (id, version). При каждом изменении баннера мы предварительно собираем все строки с заданным id, и если их количество (а это как раз количество версий баннера, которое в данный момент хранится в бд) равно 3 (или любому другому придуманному бизнесом числу), мы обновляем запись с минимальной версией. В остальных случаях (когда количество версий < 3) вместо обновления делаем INSERT
- Также задумался бы над тем, чтобы хранить баннеры в нереляционной бд. Это изначально казалось хорошей идеей, т.к. буквально с первых слов "Общих вводных": `Баннер — это документ, описывающий какой-либо элемент пользовательского интерфейса. Технически баннер представляет собой  JSON-документ неопределенной структуры.`. Т.к. баннер сам по себе является обычным JSON, хранение его в условной MongoDB кажется весьма органичным решением

### Всем добра