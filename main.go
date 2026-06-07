package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"time"

	"github.com/lelemita/who_sells_all/searcher"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

type ctxKey string
const requestIDKey ctxKey = "rid"

type ContextHandler struct {
	slog.Handler
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx != nil {
		if rid, ok := ctx.Value(requestIDKey).(string); ok {
			r.AddAttrs(slog.String("rid", rid))
		}
	}
	return h.Handler.Handle(ctx, r)
}

func initLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug, // Default to Debug level
	}
	var handler slog.Handler
	if os.Getenv("APP_ENV") == "production" {
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(&ContextHandler{Handler: handler}))
}

func main() {
	initLogger()

	// TODO ttbkey 있는지 확인하고 없으면 exit
	ttbkey := os.Getenv("ttbkey")
	if len(ttbkey) == 0 {
		slog.Error("ttbkey value is required")
		os.Exit(1)
	}
	genie := searcher.NewSearcher("https://www.aladin.co.kr", ttbkey)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		err := templates.ExecuteTemplate(w, "index.html", nil)
		if err != nil {
			slog.ErrorContext(req.Context(), "Failed to execute index.html template", slog.Any("error", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/v1/proposals", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		qry, err := url.ParseQuery(req.URL.RawQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			slog.ErrorContext(req.Context(), "Failed to parse query for proposals", slog.Any("error", err))
			fmt.Fprintf(w, `{"message": "error in ParseQuery"}`)
			return
		}
		isbns, isExist := qry["isbn"]
		if !isExist || len(isbns) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"message": "empty query isbn"}`)
			return
		}
		output := genie.GetOrderedList(req.Context(), isbns)
		jsonByte, err := json.Marshal(map[string]searcher.ShopList{"result": output})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			slog.ErrorContext(req.Context(), "Failed to marshal proposals to JSON", slog.Any("error", err))
			fmt.Fprintf(w, `{"message": "error in json.Marshal"}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(jsonByte))
	})

	http.HandleFunc("/v1/search", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		qry, err := url.ParseQuery(req.URL.RawQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			slog.ErrorContext(req.Context(), "Failed to parse query for search", slog.Any("error", err))
			fmt.Fprintf(w, `{"message": "error in ParseQuery"}`)
			return
		}
		q, isExist := qry["q"]
		if !isExist || len(q) == 0 || q[0] == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"message": "empty query q"}`)
			return
		}
		slog.InfoContext(req.Context(), "Search request", slog.String("query", q[0]))

		output, err := genie.Search(req.Context(), q[0])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			slog.WarnContext(req.Context(), "Search failed", slog.String("query", q[0]), slog.Any("error", err))
			fmt.Fprintf(w, `{"message": "%s"}`, err.Error())
			return
		}
		jsonByte, err := json.Marshal(output)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.ErrorContext(req.Context(), "Failed to marshal search result to JSON", slog.Any("error", err))
			fmt.Fprintf(w, `{"message": "error in json.Marshal"}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(jsonByte))
	})

	handler := requestIDMiddleware(loggingMiddleware(recoveryMiddleware(http.DefaultServeMux)))

	slog.Info("Starting server on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		slog.Error("Server failed to start", slog.Any("error", err))
		os.Exit(1)
	}
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (sw *statusWriter) WriteHeader(statusCode int) {
	sw.statusCode = statusCode
	sw.ResponseWriter.WriteHeader(statusCode)
}

func generateRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rid := req.Header.Get("X-Request-ID")
		if rid == "" {
			rid = generateRequestID()
		}
		ctx := context.WithValue(req.Context(), requestIDKey, rid)
		w.Header().Set("X-Request-ID", rid)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
        slog.DebugContext(req.Context(), "Request started",
            slog.String("method", req.Method),
            slog.String("path", req.URL.Path),
            slog.String("remote_addr", req.RemoteAddr),
        )
        
		sw := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(sw, req)

		slog.InfoContext(req.Context(), "Request completed",
			slog.String("method", req.Method),
			slog.String("path", req.URL.Path),
			slog.Int("status", sw.statusCode),
			slog.Duration("duration", time.Since(start)),
			slog.String("remote_addr", req.RemoteAddr),
		)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.ErrorContext(req.Context(), "Panic recovered",
					slog.Any("error", err),
					slog.String("path", req.URL.Path),
					slog.String("stack", string(debug.Stack())),
				)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"message": "error in process"}`)
			}
		}()
		next.ServeHTTP(w, req)
	})
}
