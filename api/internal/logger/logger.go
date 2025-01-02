package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"infra/api/internal/config"
	"log/slog"
	"os"
	"runtime"
	"strconv"

	pb "infra/pkg/protos/gen/go"

	"github.com/golang-cz/devslog"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Logger struct {
	pb.LogClient
}

const address = "localhost:11111"

func Init(config *config.Config) Logger {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	c := pb.NewLogClient(conn)

	slogOpts := &slog.HandlerOptions{}

	if !config.Prod_env {
		slogOpts.Level = slog.LevelDebug
	}

	// new logger with options
	opts := &devslog.Options{
		HandlerOptions:    slogOpts,
		MaxSlicePrintSize: 4,
		SortKeys:          true,
		NewLineAfterLog:   true,
	}

	logger := slog.New(devslog.NewHandler(os.Stdout, opts))

	slog.SetDefault(logger)

	return Logger{c}
}

// example Info("Coin", "Ton", "Requests", "1000")
func (l Logger) Info(message string, logStream pb.Logstream, isTemplate bool, args ...any) {
	var skip int

	const ll = LL_INFO

	if isTemplate {
		skip = 2
	} else {
		skip = 1
	}

	pc, file, line, _ := runtime.Caller(skip)
	log, err := l.formatLog(LL_INFO, message, pc, file, line, args...)
	if err != nil {
		fmt.Printf("%s:%d: format log error: %v\n", file, line, err)
		return
	}

	printLog(ll, message, file, line, args...)
	go sendLog(l.LogClient, log, logStream)
}

// example Error("Coin", "Ton", "Requests", "1000", "Error", "error text")
func (l Logger) Error(message string, logStream pb.Logstream, isTemplate bool, args ...any) {
	var skip int

	const ll = LL_ERROR

	if isTemplate {
		skip = 2
	} else {
		skip = 1
	}

	pc, file, line, _ := runtime.Caller(skip)

	log, err := l.formatLog(LL_ERROR, message, pc, file, line, args...)
	if err != nil {
		fmt.Printf("%s:%d: format log error: %v\n", file, line, err)
		return
	}

	printLog(ll, message, file, line, args...)
	go sendLog(l.LogClient, log, logStream)
}

// example Fatal("Coin", "Ton", "Requests", "1000", "Error", "error text")
func (l Logger) Fatal(message string, logStream pb.Logstream, isTemplate bool, args ...any) {
	var skip int

	const ll = LL_FATAL

	if isTemplate {
		skip = 2
	} else {
		skip = 1
	}

	pc, file, line, _ := runtime.Caller(skip)
	log, err := l.formatLog(LL_FATAL, message, pc, file, line, args...)
	if err != nil {
		fmt.Printf("%s:%d: format log error: %v\n", file, line, err)
		return
	}

	printLog(ll, message, file, line, args...)
	sendLog(l.LogClient, log, logStream)
}

func (l Logger) Debug(message string, args ...any) {
	_, file, line, _ := runtime.Caller(1)

	printLog(LL_DEBUG, message, file, line, args...)
}

func printLog(ll LogLevel, message string, file string, line int, args ...any) {
	args = append(args, "source", file+":"+strconv.Itoa(line))
	switch ll {
	case LL_ERROR:
		slog.Error(message, args...)
	case LL_INFO:
		slog.Info(message, args...)
	case LL_FATAL:
		slog.Error(message, args...)
	case LL_DEBUG:
		slog.Debug(message, args...)
	}

}

func sendLog(c pb.LogClient, buffer []byte, logstream pb.Logstream) {
	_, err := c.SendLog(context.TODO(), &pb.LogRequest{Logstream: logstream, Payload: buffer})
	if err != nil {
		fmt.Println("Error sending:", err)
		return
	}
}

func (l Logger) formatLog(ll LogLevel, message string, pc uintptr, file string, line int, args ...any) (log []byte, err error) {
	callerFunc := runtime.FuncForPC(pc).Name()

	logLevel := ll.ToString()

	logMessage := LogMessage{
		Message:  message,
		LogLevel: logLevel,
		Args:     make(map[string]interface{}),
		Source: Source{
			Function: callerFunc,
			File:     file,
			Line:     line,
		},
		AppInfo: AppInfo{
			Pid:       os.Getpid(),
			GoVersion: runtime.Version(),
		},
	}

	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			return nil, fmt.Errorf("the key must be a string: %s", args[i])
		}
		value := args[i+1]
		logMessage.Args[key] = value
	}

	b, err := json.Marshal(logMessage)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func AnyToStr(t any) string {
	return fmt.Sprintf("%v", t)
}

func GenErrorId() string {
	var errorId string
	uuid, err := uuid.NewRandom()
	if err != nil {
		errorId = NA
	} else {
		errorId = uuid.String()
	}
	return errorId
}
