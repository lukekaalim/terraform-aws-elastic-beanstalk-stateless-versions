package main

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/schema"
)

func elasticBeanstalkApplicationVersion() *schema.Resource {
	return &schema.Resource{
		Create:        createOrUpdate,
		Read:          read,
		Update:        createOrUpdate,
		Delete:        delete,
		CustomizeDiff: customizeDiff,

		Schema: map[string]*schema.Schema{
			"application_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"application_store_bucket_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"application_store_key_prefix": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "application_version_source_bundles",
			},
			"application_version_filename": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"application_version_arn": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"application_version_label": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func createOrUpdate(d *schema.ResourceData, meta interface{}) error {
	appStoreBucketName := d.Get("application_store_bucket_name").(string)
	appStoreKeyPrefix := d.Get("application_store_key_prefix").(string)
	appVersionFilename := d.Get("application_version_filename").(string)
	appName := d.Get("application_name").(string)

	appStoreVersionBundleKey := appName + "/" + appStoreKeyPrefix + "/" + appVersionFilename
	appVersionLabel := appVersionFilename[0 : len(appVersionFilename)-len(filepath.Ext(appVersionFilename))]
	d.Set("application_version_label", appVersionLabel)

	session := meta.(*session.Session)
	s3Client := s3.New(session)

	file, err := os.Open(appVersionFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(appStoreBucketName),
		Key:    aws.String(appStoreVersionBundleKey),
		Body:   bytes.NewReader(buffer),
	})
	if err != nil {
		return err
	}

	beanstalkClient := elasticbeanstalk.New(session)
	appVersionsDesc, err := beanstalkClient.DescribeApplicationVersions(&elasticbeanstalk.DescribeApplicationVersionsInput{
		ApplicationName: aws.String(appName),
		VersionLabels:   []*string{aws.String(appVersionLabel)},
	})
	if err != nil {
		return err
	}
	alreadyExists := len(appVersionsDesc.ApplicationVersions) > 0

	if alreadyExists {
		d.Set("application_version_arn", *appVersionsDesc.ApplicationVersions[0].ApplicationVersionArn)
		d.SetId(*appVersionsDesc.ApplicationVersions[0].ApplicationVersionArn)
		return read(d, meta)
	}

	versionDesc, err := beanstalkClient.CreateApplicationVersion(&elasticbeanstalk.CreateApplicationVersionInput{
		ApplicationName:       aws.String(appName),
		AutoCreateApplication: aws.Bool(false),
		Description:           aws.String("my-app-v1"),
		SourceBundle: &elasticbeanstalk.S3Location{
			S3Bucket: aws.String(appStoreBucketName),
			S3Key:    aws.String(appStoreVersionBundleKey),
		},
		VersionLabel: aws.String(appVersionLabel),
	})
	if err != nil {
		return err
	}

	d.Set("application_version_arn", *versionDesc.ApplicationVersion.ApplicationVersionArn)
	d.SetId(*versionDesc.ApplicationVersion.ApplicationVersionArn)
	return read(d, meta)
}

func customizeDiff(d *schema.ResourceDiff, m interface{}) error {
	appVersionFilename := d.Get("application_version_filename").(string)
	versionChanged := d.HasChange("application_version_filename")
	if versionChanged {
		d.SetNew("application_version_label", appVersionFilename[0:len(appVersionFilename)-len(filepath.Ext(appVersionFilename))])
		d.SetNewComputed("application_version_arn")
	}
	return nil
}

func read(d *schema.ResourceData, m interface{}) error {
	return nil
}

func delete(d *schema.ResourceData, m interface{}) error {
	return nil
}
