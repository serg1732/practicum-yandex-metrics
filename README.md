# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**


## Оптимизация памяти (pprof diff)

```text
Type: inuse_space
Time: 2026-04-22 00:28:25 MSK
Showing nodes accounting for 1859.76kB, 28.96% of 6422.23kB total
      flat  flat%   sum%        cum   cum%
 3084.03kB 48.02% 48.02%  3084.03kB 48.02%  runtime.mallocgc
-1805.17kB 28.11% 19.91% -1805.17kB 28.11%  compress/flate.NewWriter (inline)
 1584.50kB 24.67% 44.59%  1584.50kB 24.67%  compress/flate.(*dictDecoder).init (inline)
  532.26kB  8.29% 52.87%   532.26kB  8.29%  github.com/jackc/pgx/v5/pgtype.(*Map).RegisterDefaultPgType (inline)
 -516.01kB  8.03% 44.84%  -516.01kB  8.03%  github.com/jackc/pgx/v5/internal/iobufpool.init.0.func1
  515.19kB  8.02% 52.86%   515.19kB  8.02%  strings.(*Replacer).build
     514kB  8.00% 60.86%      514kB  8.00%  bufio.NewWriterSize (inline)
 -512.56kB  7.98% 52.88% -1028.57kB 16.02%  github.com/jackc/pgx/v5/pgconn.connectOne
 -512.38kB  7.98% 44.90%  -512.38kB  7.98%  github.com/jackc/pgx/v5/internal/pgio.AppendUint32 (inline)
 -512.05kB  7.97% 36.93%  -512.05kB  7.97%  context.(*cancelCtx).propagateCancel.func2
 -512.05kB  7.97% 28.96%  -512.05kB  7.97%  github.com/jackc/pgx/v5/pgconn/ctxwatch.(*ContextWatcher).Watch.func1
         0     0% 28.96%  1584.50kB 24.67%  compress/flate.NewReader
         0     0% 28.96%  1584.50kB 24.67%  compress/gzip.(*Reader).Reset
         0     0% 28.96%  1584.50kB 24.67%  compress/gzip.(*Reader).readHeader
         0     0% 28.96% -1805.17kB 28.11%  compress/gzip.(*Writer).Close
         0     0% 28.96% -1805.17kB 28.11%  compress/gzip.(*Writer).Write
         0     0% 28.96%  1584.50kB 24.67%  compress/gzip.NewReader (inline)
         0     0% 28.96%   515.19kB  8.02%  database/sql.(*DB).Ping (inline)
         0     0% 28.96%   515.19kB  8.02%  database/sql.(*DB).PingContext
         0     0% 28.96%   515.19kB  8.02%  database/sql.(*DB).PingContext.func1
         0     0% 28.96%   515.19kB  8.02%  database/sql.(*DB).conn
         0     0% 28.96%   515.19kB  8.02%  database/sql.(*DB).retry
         0     0% 28.96%   515.19kB  8.02%  database/sql.dsnConnector.Connect
         0     0% 28.96%  -512.38kB  7.98%  github.com/go-chi/chi/v5.(*Mux).Mount.func1
         0     0% 28.96%  -733.05kB 11.41%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 28.96%  -512.38kB  7.98%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 28.96%   515.19kB  8.02%  github.com/golang-migrate/migrate/v4.New
         0     0% 28.96%   515.19kB  8.02%  github.com/golang-migrate/migrate/v4/database.Open
         0     0% 28.96%   515.19kB  8.02%  github.com/golang-migrate/migrate/v4/database/postgres.(*Postgres).Open
         0     0% 28.96%   515.19kB  8.02%  github.com/golang-migrate/migrate/v4/database/postgres.WithInstance
         0     0% 28.96%  -512.38kB  7.98%  github.com/jackc/pgx/v5.(*Conn).Exec
         0     0% 28.96%  -512.38kB  7.98%  github.com/jackc/pgx/v5.(*Conn).Prepare
         0     0% 28.96%  -512.38kB  7.98%  github.com/jackc/pgx/v5.(*Conn).exec
         0     0% 28.96%  -512.38kB  7.98%  github.com/jackc/pgx/v5.(*dbTx).Exec
         0     0% 28.96%  -496.31kB  7.73%  github.com/jackc/pgx/v5.ConnectConfig
         0     0% 28.96%  -496.31kB  7.73%  github.com/jackc/pgx/v5.connect
         0     0% 28.96%  -516.01kB  8.03%  github.com/jackc/pgx/v5/internal/iobufpool.Get
         0     0% 28.96%  -512.38kB  7.98%  github.com/jackc/pgx/v5/internal/pgio.AppendInt32 (inline)
         0     0% 28.96%  -512.38kB  7.98%  github.com/jackc/pgx/v5/pgconn.(*PgConn).Prepare
         0     0% 28.96% -1028.57kB 16.02%  github.com/jackc/pgx/v5/pgconn.ConnectConfig
         0     0% 28.96%  -516.01kB  8.03%  github.com/jackc/pgx/v5/pgconn.ParseConfigWithOptions.func1
         0     0% 28.96% -1028.57kB 16.02%  github.com/jackc/pgx/v5/pgconn.connectPreferred
         0     0% 28.96%  -512.38kB  7.98%  github.com/jackc/pgx/v5/pgproto3.(*Describe).Encode
         0     0% 28.96%  -512.38kB  7.98%  github.com/jackc/pgx/v5/pgproto3.(*Frontend).SendDescribe
         0     0% 28.96%  -516.01kB  8.03%  github.com/jackc/pgx/v5/pgproto3.NewFrontend
         0     0% 28.96%  -512.38kB  7.98%  github.com/jackc/pgx/v5/pgproto3.beginMessage (inline)
         0     0% 28.96%  -516.01kB  8.03%  github.com/jackc/pgx/v5/pgproto3.newChunkReader (inline)
         0     0% 28.96%   532.26kB  8.29%  github.com/jackc/pgx/v5/pgtype.NewMap
         0     0% 28.96%   532.26kB  8.29%  github.com/jackc/pgx/v5/pgtype.initDefaultMap
         0     0% 28.96%   532.26kB  8.29%  github.com/jackc/pgx/v5/pgtype.registerDefaultPgTypeVariants[go.shape.[]github.com/jackc/pgx/v5/pgtype.Range[github.com/jackc/pgx/v5/pgtype.Float8]]
         0     0% 28.96%  -512.38kB  7.98%  github.com/jackc/pgx/v5/pgxpool.(*Tx).Exec
         0     0% 28.96%  -496.31kB  7.73%  github.com/jackc/pgx/v5/pgxpool.NewWithConfig.func3
         0     0% 28.96%  -496.31kB  7.73%  github.com/jackc/puddle/v2.(*Pool[go.shape.*uint8]).initResourceValue.func1
         0     0% 28.96%   515.19kB  8.02%  github.com/lib/pq.(*Connector).open
         0     0% 28.96%   515.19kB  8.02%  github.com/lib/pq.(*conn).auth
         0     0% 28.96%   515.19kB  8.02%  github.com/lib/pq.(*conn).startup
         0     0% 28.96%   515.19kB  8.02%  github.com/lib/pq.DialOpen
         0     0% 28.96%   515.19kB  8.02%  github.com/lib/pq.Driver.Open
         0     0% 28.96%   515.19kB  8.02%  github.com/lib/pq.Open (inline)
         0     0% 28.96%   515.19kB  8.02%  github.com/lib/pq/scram.(*Client).Step
         0     0% 28.96%   515.19kB  8.02%  github.com/lib/pq/scram.(*Client).step1
         0     0% 28.96% -1805.17kB 28.11%  github.com/serg1732/practicum-yandex-metrics/internal/helpers/compress.(*compressWriter).Close
         0     0% 28.96%  1584.50kB 24.67%  github.com/serg1732/practicum-yandex-metrics/internal/helpers/compress.NewCompressReader
         0     0% 28.96%  -512.38kB  7.98%  github.com/serg1732/practicum-yandex-metrics/internal/repository.(*DataBase).Update
         0     0% 28.96%  -512.38kB  7.98%  github.com/serg1732/practicum-yandex-metrics/internal/repository.(*DataBase).Update.func1
         0     0% 28.96%  -512.38kB  7.98%  github.com/serg1732/practicum-yandex-metrics/internal/repository.(*DataBase).retry
         0     0% 28.96%   515.19kB  8.02%  github.com/serg1732/practicum-yandex-metrics/internal/repository.MigrateDataBase
         0     0% 28.96%  -733.05kB 11.41%  main.buildRouter.WithCheckHash.func6.1
         0     0% 28.96%  -733.05kB 11.41%  main.buildRouter.WithGzipCompress.func7.1
         0     0% 28.96%  -733.05kB 11.41%  main.buildRouter.WithLogger.func5.1
         0     0% 28.96%  -733.05kB 11.41%  main.buildRouter.func1.1
         0     0% 28.96%  -512.38kB  7.98%  main.buildRouter.func3.(*UpdateHandlerImpl).UpdateJSONHandler.1
         0     0% 28.96%   515.19kB  8.02%  main.main
         0     0% 28.96%  -219.05kB  3.41%  net/http.(*conn).serve
         0     0% 28.96%  -733.05kB 11.41%  net/http.HandlerFunc.ServeHTTP
         0     0% 28.96%      514kB  8.00%  net/http.newBufioWriterSize
         0     0% 28.96%  -733.05kB 11.41%  net/http.serverHandler.ServeHTTP
         0     0% 28.96%     2565kB 39.94%  runtime.allocm
         0     0% 28.96%   515.19kB  8.02%  runtime.main
         0     0% 28.96%     2565kB 39.94%  runtime.mstart
         0     0% 28.96%     2565kB 39.94%  runtime.mstart0
         0     0% 28.96%     2565kB 39.94%  runtime.mstart1
         0     0% 28.96%     2565kB 39.94%  runtime.newm
         0     0% 28.96%  3084.03kB 48.02%  runtime.newobject
         0     0% 28.96%   519.03kB  8.08%  runtime.procresize
         0     0% 28.96%     2565kB 39.94%  runtime.resetspinning
         0     0% 28.96%   519.03kB  8.08%  runtime.rt0_go
         0     0% 28.96%   519.03kB  8.08%  runtime.schedinit
         0     0% 28.96%     2565kB 39.94%  runtime.schedule
         0     0% 28.96%     2565kB 39.94%  runtime.startm
         0     0% 28.96%     2565kB 39.94%  runtime.wakep
         0     0% 28.96%   515.19kB  8.02%  strings.(*Replacer).WriteString
         0     0% 28.96%   515.19kB  8.02%  strings.(*Replacer).buildOnce
         0     0% 28.96%  1047.45kB 16.31%  sync.(*Once).Do (inline)
         0     0% 28.96%  1047.45kB 16.31%  sync.(*Once).doSlow
         0     0% 28.96%  -516.01kB  8.03%  sync.(*Pool).Get
```