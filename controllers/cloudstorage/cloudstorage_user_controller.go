package cloudstorage

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"solana/utilities"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/gin-gonic/gin"
)

type CloudstorageController struct {
	s3Client *s3.Client
}

var cc *CloudstorageController

func Init() {
	cfgData := getS3Config()

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfgData["AccessKeyId"],
			cfgData["SecretAccessKey"],
			"",
		)),
		config.WithRegion(cfgData["Region"]),
		config.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               cfgData["Endpoint"],
				HostnameImmutable: true,
			}, nil
		})),
	)
	if err != nil {
		panic(fmt.Sprintf("Unable to load AWS SDK config: %v", err))
	}

	s3Client := s3.NewFromConfig(awsCfg)
	cc = &CloudstorageController{s3Client: s3Client}
}

func getS3Config() map[string]string {
	return map[string]string{
		"AccessKeyId":     utilities.Get[string]("cf_access_id"),
		"SecretAccessKey": utilities.Get[string]("cf_secret_key"),
		"Region":          utilities.Get[string]("cf_region"),
		"Endpoint":        utilities.Get[string]("cf_endpoint"),
	}
}

func GetUsersCloudstorageFiles2(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	accountId := c.Param("accountId")
	ua := utilities.Parse(c.GetHeader("User-Agent"))
	if ua == nil {
		utilities.Basic.BadRequest().Apply(c.Writer)
		return
	}
	seasonStr := strconv.Itoa(ua.Season)
	s3Key := fmt.Sprintf("Settings/%s/ClientSettings-%s.Sav", accountId, seasonStr)

	input := &s3.GetObjectInput{
		Bucket: aws.String("solaris"),
		Key:    aws.String(s3Key),
	}
	result, err := cc.s3Client.GetObject(context.TODO(), input)
	if err != nil {
		c.JSON(http.StatusOK, []gin.H{})
		return
	}
	defer result.Body.Close()

	fileBytes, err := io.ReadAll(result.Body)
	if err != nil {
		c.JSON(http.StatusOK, []gin.H{})
		return
	}

	sha1Hash := sha1.Sum(fileBytes)
	sha256Hash := sha256.Sum256(fileBytes)
	fileInfo := []gin.H{
		{
			"uniqueFilename": "ClientSettings.Sav",
			"filename":       "ClientSettings.Sav",
			"hash":           hex.EncodeToString(sha1Hash[:]),
			"hash256":        hex.EncodeToString(sha256Hash[:]),
			"length":         len(fileBytes),
			"contentType":    "application/octet-stream",
			"uploaded":       result.LastModified,
			"storageType":    "S3",
			"accountId":      accountId,
		},
	}
	c.JSON(http.StatusOK, fileInfo)
}

func GetUserCloudstorageFile(c *gin.Context) {
	accountId := c.Param("accountId")
	filename := c.Param("filename")
	if filename != "ClientSettings.Sav" && filename != "ClientSettingsIos.Sav" {
		c.Status(http.StatusOK)
		return
	}
	ua := utilities.Parse(c.GetHeader("User-Agent"))
	if ua == nil {
		utilities.Basic.BadRequest().Apply(c.Writer)
		return
	}
	seasonStr := strconv.Itoa(ua.Season)
	s3Key := fmt.Sprintf("Settings/%s/ClientSettings-%s.Sav", accountId, seasonStr)

	input := &s3.GetObjectInput{
		Bucket: aws.String("solaris"),
		Key:    aws.String(s3Key),
	}
	result, err := cc.s3Client.GetObject(context.TODO(), input)
	if err != nil {
		fmt.Printf("Error fetching S3 object: %v\n", err)
		c.Status(http.StatusOK)
		return
	}
	defer result.Body.Close()

	c.Header("Content-Type", "application/octet-stream")
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, result.Body); err != nil {
		fmt.Printf("Error reading S3 object body: %v\n", err)
		c.Status(http.StatusOK)
		return
	}
	c.Data(http.StatusOK, "application/octet-stream", buf.Bytes())
}

func SaveUsersCloudstorageFile(c *gin.Context) {
	accountId := c.Param("accountId")
	filename := c.Param("filename")

	if filename != "ClientSettings.Sav" && filename != "ClientSettingsIos.Sav" {
		c.Status(http.StatusBadRequest)
		return
	}

	ua := utilities.Parse(c.GetHeader("User-Agent"))
	if ua == nil {
		utilities.Basic.BadRequest().Apply(c.Writer)
		return
	}

	seasonStr := strconv.Itoa(ua.Season)

	fileBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	s3Key := fmt.Sprintf("Settings/%s/%s-%s.Sav", accountId, strings.TrimSuffix(filename, ".Sav"), seasonStr)

	input := &s3.PutObjectInput{
		Bucket:        aws.String("solaris"),
		Key:           aws.String(s3Key),
		Body:          bytes.NewReader(fileBytes),
		ContentLength: aws.Int64(int64(len(fileBytes))),
		ContentType:   aws.String("application/octet-stream"),
	}

	_, err = cc.s3Client.PutObject(context.TODO(), input)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
