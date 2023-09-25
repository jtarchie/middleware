package middleware

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
)

// Logger is a middleware and slog to provide an "access log" like logging for each request.
func Logger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			start := time.Now()

			req := context.Request()

			requestID := req.Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = uuid.Must(uuid.NewV7()).String()
				req.Header.Set(echo.HeaderXRequestID, requestID)
			}

			err := next(context)
			if err != nil {
				context.Error(err)
			}

			stop := time.Now()
			res := context.Response()
			res.Header().Set(echo.HeaderXRequestID, requestID)

			contentLength := req.Header.Get(echo.HeaderContentLength)
			if contentLength == "" {
				contentLength = "0"
			}

			latency := stop.Sub(start)

			const preferredBase = 10
			fields := []any{
				slog.String("bytes_in", contentLength),
				slog.String("bytes_out", strconv.FormatInt(res.Size, preferredBase)),
				slog.String("host", req.Host),
				slog.String("id", requestID),
				slog.String("latency_human", stop.Sub(start).String()),
				slog.String("latency", strconv.FormatInt(int64(latency), preferredBase)),
				slog.String("method", req.Method),
				slog.String("remote_ip", context.RealIP()),
				slog.Int("status", res.Status),
				slog.String("time", time.Now().Format(time.RFC3339Nano)),
				slog.String("uri", req.RequestURI),
				slog.String("user_agent", req.UserAgent()),
			}

			if err != nil {
				fields = append(fields, slog.String("error", err.Error()))
			}

			logger.Info("http_request", fields...)

			return nil
		}
	}
}
