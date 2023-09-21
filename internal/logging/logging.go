package logging

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var logger *zap.SugaredLogger

func Init() error {
	var err error
	z, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("logger don't Run! %s", err)
	}
	logger = z.Sugar()
	return nil
}

func Infof(format string, args ...any) {
	logger.Infof(format, args...)
}

func Fatalf(format string, args ...any) {
	logger.Fatalf(format, args...)
}

func Errorf(format string, args ...any) {
	logger.Errorf(format, args...)
}

func CustomMiddlewareLogger(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)

		logger.Infoln("URI:", r.RequestURI,
			"Method:", r.Method,
			"Duration:", time.Since(start))
	}

	return http.HandlerFunc(fn)
}
