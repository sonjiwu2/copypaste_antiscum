# Groq Cloud Free на Render

Ключ добавляется только в backend Web Service. В образ и frontend его
встраивать нельзя.

1. Откройте в Render сервис backend.
2. Перейдите в **Environment**.
3. Добавьте Secret/Environment Variable `GROQ_API_KEY` со значением нового
   ключа из `https://console.groq.com/keys`.
4. Добавьте обычные переменные:
   - `GROQ_MODEL=openai/gpt-oss-20b`
   - `GROQ_BASE_URL=https://api.groq.com/openai/v1`
   - `GROQ_REQUEST_TIMEOUT=40s`
   - `HTTP_WRITE_TIMEOUT=60s`
5. Сохраните изменения и выберите **Manual Deploy → Deploy latest image**.

Для сохранения одного теста на неделю backend должен работать с PostgreSQL и
миграциями `00005_weekly_tests.sql`, `00006_weekly_test_groq_source.sql` и
`00007_weekly_test_exam.sql`.
Перед запуском нового backend-образа
выполните `/usr/local/bin/migrate up` как pre-deploy command или отдельный job.
Если используется `STORAGE_DRIVER=memory`, Groq тоже работает, но после
перезапуска сервиса тест забудется и следующий запрос снова потратит лимит.

Проверка после deploy:

```text
POST https://<backend>.onrender.com/api/v1/weekly-tests/current
```

В браузере это проще проверить кнопкой «Пройти тест» во frontend. Поле
`source` в ответе равно `groq`, если Groq ответил успешно, и `fallback`, если
ключ отсутствует, бесплатный лимит исчерпан или API временно недоступен.

Новый формат — 20 вопросов, 10 минут, 3 жизни, проходной результат 18/20 и
100 XP только за успешную сдачу. Следующий тест создаётся через 7 суток после
завершения предыдущего. Старые результаты тестов из 5 вопросов миграция
сохраняет, но для нового формата создаётся отдельный экзамен.

Ключ храните только в секрете backend. Не добавляйте его во frontend, Docker
image, `VITE_*`-переменные или Git. Бесплатный тариф имеет ограничения по
частоте и числу запросов; превышение лимита не ломает демонстрацию благодаря
резервному тесту.
