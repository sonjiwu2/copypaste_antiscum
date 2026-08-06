package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
)

// requestHeaderName — заголовок, через который клиент или прокси может задать
// свой идентификатор запроса для сквозной трассировки.
const requestHeaderName = "X-Request-ID"

// statusRecorder запоминает код ответа для логирования.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

func (s *statusRecorder) Write(body []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}

	return s.ResponseWriter.Write(body)
}

// withRequestContext присваивает запросу идентификатор и кладёт в контекст
// логгер, уже помеченный этим идентификатором.
func withRequestContext(logger *slog.Logger, generator identifier.Generator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(requestHeaderName)
		if requestID == "" {
			generated, err := generator.NewID()
			if err != nil {
				logger.ErrorContext(r.Context(), "не удалось сгенерировать идентификатор запроса",
					slog.String("error", err.Error()))
			}

			requestID = generated
		}

		w.Header().Set(requestHeaderName, requestID)

		requestLogger := logger.With(slog.String("requestId", requestID))
		ctx := withLogger(withRequestID(r.Context(), requestID), requestLogger)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// withAccessLog пишет одну запись на завершённый запрос.
func withAccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		if recorder.status == 0 {
			recorder.status = http.StatusOK
		}

		loggerFrom(r.Context()).InfoContext(r.Context(), "запрос обработан",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", recorder.status),
			slog.Duration("duration", time.Since(startedAt)),
		)
	})
}

// withPanicRecovery не даёт панике уронить процесс и скрывает stack trace от клиента.
func withPanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			loggerFrom(r.Context()).ErrorContext(r.Context(), "паника при обработке запроса",
				slog.Any("panic", recovered))

			writeError(w, r, http.StatusInternalServerError, CodeInternalError,
				"Внутренняя ошибка сервера.")
		}()

		next.ServeHTTP(w, r)
	})
}

// withBodyLimit ограничивает размер тела запроса.
func withBodyLimit(maxBytes int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		}

		next.ServeHTTP(w, r)
	})
}
