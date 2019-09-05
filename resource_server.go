package main

import (
	"bytes"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCreate,
		Read:   resourceServerRead,
		Update: resourceServerCreate,
		Delete: resourceServerDelete,

		Schema: map[string]*schema.Schema{
			"application_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"application_store_bucket_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"application_version_filename": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceServerCreate(d *schema.ResourceData, meta interface{}) error {
	appStoreBucketName := d.Get("application_store_bucket_name").(string)
	appVersionFilename := d.Get("application_version_filename").(string)

	session := meta.(*session.Session)
	s3Client := s3.New(session)

	file, _ := os.Open(appVersionFilename)
	defer file.Close()

	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(appStoreBucketName),
		Key:    aws.String("Howdy_Partner"),
		Body:   bytes.NewReader(buffer),
	})

	//beanstalkClient := elasticbeanstalk.New(session)
	/*
		beanstalkClient.CreateApplicationVersion(&elasticbeanstalk.CreateApplicationVersionInput{
			ApplicationName:       aws.String("my-app"),
			AutoCreateApplication: aws.Bool(true),
			Description:           aws.String("my-app-v1"),
			Process:               aws.Bool(true),
			SourceBundle: &elasticbeanstalk.S3Location{
				S3Bucket: aws.String(appStoreBucketName),
				S3Key:    aws.String("Howdy_Partner"),
			},
			VersionLabel: aws.String("v1"),
		})
	*/

	applicationName := d.Get("application_name").(string)
	d.SetId(applicationName)
	return resourceServerRead(d, meta)
}

func resourceServerRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceServerUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceServerRead(d, m)
}

func resourceServerDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
