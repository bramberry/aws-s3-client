package simplerest

import (
	"bytes"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/globalsign/mgo/bson"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

var (
	envNotFound  = errors.New("environment variable not found")
	awsSecretId  = "AWS_SECRET_ID"
	awsSecretKey = "AWS_SECRET_KEY"
	buf50MB      = make([]byte, 1024*1024*50)
)

// Create S3 service client
func InitAWSClient() (*session.Session, error) {
	// create an AWS session which can be reused
	key, exists := os.LookupEnv(awsSecretId)
	if !exists {
		return nil, envNotFound
	}
	secret, exists := os.LookupEnv(awsSecretKey)
	if !exists {
		return nil, envNotFound
	}
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(s3.BucketLocationConstraintEuCentral1),
		Credentials: credentials.NewStaticCredentials(key, secret, ""),
	})

	return sess, err
}

// UploadFileToS3 saves a file to aws bucket and returns the url to // the file and an error if there's any
func UploadFileToS3(s *server, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	// get the file size and read
	// the file content into a buffer
	size := fileHeader.Size
	buffer := make([]byte, size)
	_, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	// create a unique file name for the file
	tempFileName := Config.AWSPicturesFolderName + bson.NewObjectId().Hex() + filepath.Ext(fileHeader.Filename)

	_, err = s3.New(s.awsSession).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(Config.AWSBucketName),
		Key:                  aws.String(tempFileName),
		ACL:                  aws.String(s3.BucketCannedACLPublicRead),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String(s3.ServerSideEncryptionAes256),
		StorageClass:         aws.String(s3.StorageClassStandard),
	})

	return tempFileName, err
}

// DownloadFileToS3 download file from aws bucket
func DownloadFileToS3(s *server, filePath string) (*aws.WriteAtBuffer, error) {
	file := aws.NewWriteAtBuffer(make([]byte, len(buf50MB)))
	downloader := s3manager.NewDownloader(s.awsSession)

	_, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(Config.AWSBucketName),
			Key:    aws.String(filePath),
		})

	return file, err
}
