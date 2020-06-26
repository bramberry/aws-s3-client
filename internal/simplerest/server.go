package simplerest

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"time"
)

const (
	sessionName = "gopherschool"
	ctxKeyRequestID
)

type server struct {
	router     *mux.Router
	logger     *logrus.Logger
	awsSession *session.Session
}

func newServer() *server {
	s := &server{
		logger: logrus.New(),
		router: mux.NewRouter(),
	}
	sess, err := InitAWSClient()
	if err != nil {
		s.logger.Errorf("Failed to create AWS session: %v", err)
	}
	s.awsSession = sess
	s.configureRouter()
	return s
}

func (s *server) configureRouter() {
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	s.router.HandleFunc("/upload", s.uploadHandler()).Methods(http.MethodPost)
	s.router.HandleFunc("/download", s.downloadHandler()).Methods(http.MethodGet)
}

func (s *server) uploadHandler() func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		maxSize := int64(1024000) // allow only 1MB of file size
		err := request.ParseMultipartForm(maxSize)
		if err != nil {
			log.Println(err)
			fmt.Fprintf(writer, "Image too large. Max Size: %v", maxSize)
			return
		}
		file, fileHeader, err := request.FormFile("file")
		if err != nil {
			log.Println(err)
			fmt.Fprintf(writer, "Could not get uploaded file")
			return
		}
		defer file.Close()

		fileName, err := UploadFileToS3(s, file, fileHeader)
		if err != nil {
			fmt.Fprintf(writer, "Could not upload file")
		}
		fmt.Fprintf(writer, "Image uploaded successfully: %v", fileName)
	}
}

func (s *server) downloadHandler() func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		path := request.Form.Get("path")
		file, err := DownloadFileToS3(s, path)
		if err != nil {
			fmt.Fprintf(writer, "Could not download file %s from S3", path)
		}
		_, err = writer.Write(file.Bytes())
		if err != nil {
			fmt.Fprint(writer, "Could not write file")
		}
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(ctxKeyRequestID),
		})
		logger.Infof("started %s %s", r.Method, r.RequestURI)

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		var level logrus.Level
		switch {
		case rw.code >= 500:
			level = logrus.ErrorLevel
		case rw.code >= 400:
			level = logrus.WarnLevel
		default:
			level = logrus.InfoLevel
		}
		logger.Logf(
			level,
			"completed with %d %s in %v",
			rw.code,
			http.StatusText(rw.code),
			time.Now().Sub(start),
		)
	})
}
