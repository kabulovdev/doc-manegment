package s3compat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
)

type Config struct {
	Endpoint       string
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	ForcePathStyle bool
	HTTPClient     *http.Client
}

type Provider struct {
	client *s3.Client
	bucket string
}

func New(ctx context.Context, cfg Config) (*Provider, error) {
	region := cfg.Region
	if region == "" {
		region = "auto"
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		awsconfig.WithHTTPClient(cfg.HTTPClient),
		awsconfig.WithRequestChecksumCalculation(aws.RequestChecksumCalculationWhenRequired),
		awsconfig.WithResponseChecksumValidation(aws.ResponseChecksumValidationWhenRequired),
	)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.ForcePathStyle
	})
	return &Provider{client: client, bucket: cfg.Bucket}, nil
}

func NewFromStorageConfig(ctx context.Context, sc *domain.StorageConfig, accessKey, secretKey string, httpClient *http.Client) (*Provider, error) {
	return New(ctx, Config{
		Endpoint:       sc.Endpoint,
		Region:         sc.Region,
		Bucket:         sc.Bucket,
		AccessKey:      accessKey,
		SecretKey:      secretKey,
		ForcePathStyle: sc.ForcePathStyle,
		HTTPClient:     httpClient,
	})
}

func (p *Provider) HeadBucket(ctx context.Context) error {
	_, err := p.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(p.bucket)})
	return err
}

func (p *Provider) Stat(ctx context.Context, key string) (ports.ObjectInfo, error) {
	out, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return ports.ObjectInfo{}, mapErr(err)
	}
	info := ports.ObjectInfo{Key: key}
	if out.ContentLength != nil {
		info.Size = *out.ContentLength
	}
	if out.ETag != nil {
		info.ETag = strings.Trim(*out.ETag, `"`)
	}
	if out.ContentType != nil {
		info.ContentType = *out.ContentType
	}
	if out.LastModified != nil {
		info.LastModified = out.LastModified.Unix()
	}
	return info, nil
}

func (p *Provider) GetStream(ctx context.Context, key string, rng *ports.Range) (io.ReadCloser, ports.ObjectInfo, error) {
	in := &s3.GetObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key)}
	if rng != nil {
		if rng.End < 0 {
			in.Range = aws.String(fmt.Sprintf("bytes=%d-", rng.Start))
		} else {
			in.Range = aws.String(fmt.Sprintf("bytes=%d-%d", rng.Start, rng.End))
		}
	}
	out, err := p.client.GetObject(ctx, in)
	if err != nil {
		return nil, ports.ObjectInfo{}, mapErr(err)
	}
	info := ports.ObjectInfo{Key: key}
	if out.ContentLength != nil {
		info.Size = *out.ContentLength
	}
	if out.ETag != nil {
		info.ETag = strings.Trim(*out.ETag, `"`)
	}
	if out.ContentType != nil {
		info.ContentType = *out.ContentType
	}
	return out.Body, info, nil
}

func (p *Provider) PutStream(ctx context.Context, key string, body io.Reader, size int64, mime string) error {
	in := &s3.PutObjectInput{
		Bucket:        aws.String(p.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
	}
	if mime != "" {
		in.ContentType = aws.String(mime)
	}
	_, err := p.client.PutObject(ctx, in)
	return mapErr(err)
}

func (p *Provider) Delete(ctx context.Context, key string) error {
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	return mapErr(err)
}

func (p *Provider) List(ctx context.Context, prefix, cursor string) (ports.ListPage, error) {
	in := &s3.ListObjectsV2Input{
		Bucket:  aws.String(p.bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(1000),
	}
	if cursor != "" {
		in.ContinuationToken = aws.String(cursor)
	}
	out, err := p.client.ListObjectsV2(ctx, in)
	if err != nil {
		return ports.ListPage{}, mapErr(err)
	}
	page := ports.ListPage{IsTruncated: out.IsTruncated != nil && *out.IsTruncated}
	if out.NextContinuationToken != nil {
		page.NextCursor = *out.NextContinuationToken
	}
	for _, o := range out.Contents {
		info := ports.ObjectInfo{}
		if o.Key != nil {
			info.Key = *o.Key
		}
		if o.Size != nil {
			info.Size = *o.Size
		}
		if o.ETag != nil {
			info.ETag = strings.Trim(*o.ETag, `"`)
		}
		if o.LastModified != nil {
			info.LastModified = o.LastModified.Unix()
		}
		page.Objects = append(page.Objects, info)
		page.TotalBytes += info.Size
		page.TotalCount++
	}
	return page, nil
}

func (p *Provider) InitiateMultipart(ctx context.Context, key, mime string) (string, error) {
	in := &s3.CreateMultipartUploadInput{Bucket: aws.String(p.bucket), Key: aws.String(key)}
	if mime != "" {
		in.ContentType = aws.String(mime)
	}
	out, err := p.client.CreateMultipartUpload(ctx, in)
	if err != nil {
		return "", mapErr(err)
	}
	if out.UploadId == nil {
		return "", errors.New("no upload id")
	}
	return *out.UploadId, nil
}

func (p *Provider) UploadPart(ctx context.Context, key, uploadID string, partNum int, body io.Reader, size int64) (string, error) {
	out, err := p.client.UploadPart(ctx, &s3.UploadPartInput{
		Bucket:        aws.String(p.bucket),
		Key:           aws.String(key),
		UploadId:      aws.String(uploadID),
		PartNumber:    aws.Int32(int32(partNum)),
		Body:          body,
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return "", mapErr(err)
	}
	if out.ETag == nil {
		return "", errors.New("no etag")
	}
	return strings.Trim(*out.ETag, `"`), nil
}

func (p *Provider) CompleteMultipart(ctx context.Context, key, uploadID string, parts []ports.Part) error {
	mpu := make([]s3types.CompletedPart, 0, len(parts))
	for _, pt := range parts {
		mpu = append(mpu, s3types.CompletedPart{
			PartNumber: aws.Int32(int32(pt.PartNumber)),
			ETag:       aws.String(`"` + pt.ETag + `"`),
		})
	}
	_, err := p.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(p.bucket),
		Key:             aws.String(key),
		UploadId:        aws.String(uploadID),
		MultipartUpload: &s3types.CompletedMultipartUpload{Parts: mpu},
	})
	return mapErr(err)
}

func (p *Provider) AbortMultipart(ctx context.Context, key, uploadID string) error {
	_, err := p.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(p.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	})
	return mapErr(err)
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	var nsk *s3types.NoSuchKey
	if errors.As(err, &nsk) {
		return domainNotFound(err)
	}
	var nf *s3types.NotFound
	if errors.As(err, &nf) {
		return domainNotFound(err)
	}
	return err
}

func domainNotFound(_ error) error { return domain.ErrNotFound }
