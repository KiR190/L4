package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	logCh   = make(chan string, 256)
	once    sync.Once
	closeCh = make(chan struct{})
	wg      sync.WaitGroup
)

func init() {
	startLogger()
}

func startLogger() {
	once.Do(func() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-closeCh:
					// Обрабатываем оставшиеся сообщения
					for len(logCh) > 0 {
						msg := <-logCh
						log.Print(msg)
					}
					return
				case msg := <-logCh:
					log.Print(msg)
				}
			}
		}()
	})
}

// Close корректно завершает работу логгера
func Close() {
	close(closeCh)
	wg.Wait()
}

// Info логирует информационное сообщение асинхронно
func Info(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	select {
	case logCh <- msg:
	default:
		// Если канал переполнен, пишем напрямую
		log.Print(msg)
	}
}

// Error логирует ошибку асинхронно
func Error(format string, v ...interface{}) {
	msg := fmt.Sprintf("ERROR: "+format, v...)
	select {
	case logCh <- msg:
	default:
		log.Print(msg)
	}
}

// Printf
func Printf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	select {
	case logCh <- msg:
	default:
		log.Print(msg)
	}
}

// Println
func Println(v ...interface{}) {
	msg := fmt.Sprint(v...)
	select {
	case logCh <- msg:
	default:
		log.Print(msg)
	}
}

// Fatalf логирует критическую ошибку и завершает программу
func Fatalf(format string, v ...interface{}) {
	msg := fmt.Sprintf("FATAL: "+format, v...)
	// Для fatal сообщений пишем напрямую и завершаем
	log.Print(msg)
	Close()
	os.Exit(1)
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		msg := fmt.Sprintf("%s | %s %s | %d | %v",
			time.Now().Format("2006/01/02 - 15:04:05"),
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
		)
		select {
		case logCh <- msg:
		default:
			log.Print(msg)
		}
	}
}
