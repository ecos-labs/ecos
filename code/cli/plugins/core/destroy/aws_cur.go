package destroy

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	cliConfig "github.com/ecos-labs/ecos-core/code/cli/config"
	"github.com/ecos-labs/ecos-core/code/cli/plugins/registry"
	"github.com/ecos-labs/ecos-core/code/cli/plugins/types"
	"github.com/ecos-labs/ecos-core/code/cli/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
)

type AwsCurDestroyPlugin struct {
	region         string
	bucket         string
	dbtWorkgroup   string
	adhocWorkgroup string
	awsProfile     string
	accountID      string
}

func NewAwsCurDestroy() types.DestroyPlugin {
	return &AwsCurDestroyPlugin{}
}

// Self-register the plugin
func init() {
	registry.RegisterDestroyPlugin("aws_cur", NewAwsCurDestroy)
}

func (p *AwsCurDestroyPlugin) Name() string { return "aws_cur" }

func (p *AwsCurDestroyPlugin) LoadFromConfig(cfg *cliConfig.EcosConfig) error {
	if cfg == nil {
		return errors.New("nil config")
	}

	p.region = cfg.AWS.Region
	p.bucket = cfg.AWS.ResultsBucket
	p.dbtWorkgroup = cfg.AWS.DBTWorkgroup
	p.adhocWorkgroup = cfg.AWS.AdhocWorkgroup

	p.awsProfile = cfg.Transform.DBT.AWSProfile

	if p.region == "" {
		return errors.New("aws.region missing in .ecos.yaml")
	}
	noResources := p.bucket == "" &&
		p.dbtWorkgroup == "" &&
		p.adhocWorkgroup == ""

	if noResources {
		return errors.New(
			`.ecos.yaml does not contain any ecos-managed resource names.

This usually happens when:
  • Resource creation was skipped during "ecos init", or
  • The .ecos.yaml file was manually modified or corrupted.

Please re-run "ecos init" (or review your existing .ecos.yaml) before running "ecos destroy"`,
		)
	}

	return nil
}

func (p *AwsCurDestroyPlugin) DescribeDestruction() []types.DestroyResourcePreview {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(p.region)}
	if p.awsProfile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(p.awsProfile))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		// If we can't load AWS config, return error preview for all resources
		return []types.DestroyResourcePreview{
			{
				Kind:    "S3 Bucket",
				Name:    p.bucket,
				Managed: false,
				Error:   humanizePreviewError(err, ""),
			},
			{
				Kind:    "Workgroup",
				Name:    p.dbtWorkgroup,
				Managed: false,
				Error:   humanizePreviewError(err, ""),
			},
		}
	}

	s3Client := s3.NewFromConfig(cfg)
	athClient := athena.NewFromConfig(cfg)

	results := []types.DestroyResourcePreview{}

	// 1. Bucket tag check
	bucketPreview := types.DestroyResourcePreview{
		Kind: "S3 Bucket",
		Name: p.bucket,
	}
	if managedBucket, err := p.isBucketManaged(ctx, s3Client, p.bucket); err != nil {
		bucketPreview.Error = humanizePreviewError(err, p.bucket)
	} else {
		bucketPreview.Managed = managedBucket
	}
	results = append(results, bucketPreview)

	// 2. DBT workgroup tag check
	dbtPreview := types.DestroyResourcePreview{
		Kind: "Workgroup",
		Name: p.dbtWorkgroup,
	}
	if managedDBT, err := p.isWorkgroupManaged(ctx, athClient, p.dbtWorkgroup); err != nil {
		dbtPreview.Error = humanizePreviewError(err, p.dbtWorkgroup)
	} else {
		dbtPreview.Managed = managedDBT
	}
	results = append(results, dbtPreview)

	// 3. Adhoc workgroup tag check
	if p.adhocWorkgroup != "" {
		adhocPreview := types.DestroyResourcePreview{
			Kind: "Workgroup",
			Name: p.adhocWorkgroup,
		}
		if managedAdhoc, err := p.isWorkgroupManaged(ctx, athClient, p.adhocWorkgroup); err != nil {
			adhocPreview.Error = humanizePreviewError(err, p.adhocWorkgroup)
		} else {
			adhocPreview.Managed = managedAdhoc
		}
		results = append(results, adhocPreview)
	}

	return results
}

func (p *AwsCurDestroyPlugin) ValidatePrerequisites() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(p.region),
	}
	if p.awsProfile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(p.awsProfile))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return fmt.Errorf("LoadDefaultConfig failed: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	ident, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("STS GetCallerIdentity failed: %w", err)
	}

	p.accountID = *ident.Account
	return nil
}

func (p *AwsCurDestroyPlugin) DestroyResources() ([]types.DestroyResourceResult, error) {
	ctx := context.Background()

	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(p.region),
	}
	if p.awsProfile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(p.awsProfile))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("LoadDefaultConfig failed: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	athClient := athena.NewFromConfig(cfg)

	var results []types.DestroyResourceResult
	isEmpty, err := p.isBucketEmpty(ctx, s3Client, p.bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket is empty: %w", err)
	}

	if !isEmpty {
		utils.PrintWarning(fmt.Sprintf(
			"The S3 bucket '%s' is not empty and contains objects.\nDeleting it will permanently remove all data.",
			p.bucket,
		))

		confirm := utils.ConfirmPrompt("Continue deleting this bucket")
		if !confirm {
			return nil, nil
		}
	}

	spinner := utils.NewSpinner("Destroying aws_cur resources...")
	spinner.Start()

	// Bucket
	bucketRes := p.destroyBucket(ctx, s3Client, p.bucket)
	results = append(results, bucketRes)

	// DBT WG
	if p.dbtWorkgroup != "" {
		wgRes := p.destroyWorkgroup(ctx, athClient, p.dbtWorkgroup)
		results = append(results, wgRes)
	}

	// Adhoc WG
	if p.adhocWorkgroup != "" {
		wgRes := p.destroyWorkgroup(ctx, athClient, p.adhocWorkgroup)
		results = append(results, wgRes)
	}

	hasFailure := false
	for _, r := range results {
		if r.Status == types.DestroyStatusFailed {
			hasFailure = true
			break
		}
	}

	if hasFailure {
		spinner.Error("Destruction failed")
		return results, errors.New("one or more resources failed to destroy")
	}

	spinner.Success("aws_cur resources destroyed successfully")
	return results, nil
}

func (p *AwsCurDestroyPlugin) isBucketManaged(ctx context.Context, client *s3.Client, bucket string) (bool, error) {
	tagRes, err := client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{Bucket: &bucket})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchTagSet") ||
			strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, err
	}

	for _, t := range tagRes.TagSet {
		key := aws.ToString(t.Key)
		value := aws.ToString(t.Value)
		if key == "ecos:managed" && value == "true" {
			return true, nil
		}
	}
	return false, nil
}

func (p *AwsCurDestroyPlugin) isWorkgroupManaged(ctx context.Context, client *athena.Client, wg string) (bool, error) {
	wgARN := fmt.Sprintf("arn:aws:athena:%s:%s:workgroup/%s", p.region, p.accountID, wg)

	tagRes, err := client.ListTagsForResource(ctx, &athena.ListTagsForResourceInput{
		ResourceARN: &wgARN,
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchTagSet") {
			return false, nil
		}
		return false, err
	}

	for _, t := range tagRes.Tags {
		key := aws.ToString(t.Key)
		value := aws.ToString(t.Value)
		if key == "ecos:managed" && value == "true" {
			return true, nil
		}
	}
	return false, nil
}

func (p *AwsCurDestroyPlugin) destroyBucket(
	ctx context.Context,
	client *s3.Client,
	bucket string,
) types.DestroyResourceResult {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &bucket,
	})
	if err != nil {
		return types.DestroyResourceResult{
			Kind:   "S3 Bucket",
			Name:   bucket,
			Status: types.DestroyStatusSkipped,
			Error:  "Bucket already deleted or does not exist",
		}
	}

	// Delete versions first (if any)
	if err := p.deleteObjectVersions(ctx, client, bucket); err != nil {
		return types.DestroyResourceResult{
			Kind:   "S3 Bucket",
			Name:   bucket,
			Status: types.DestroyStatusFailed,
			Error:  fmt.Sprintf("failed to delete object versions: %v", err),
		}
	}

	// Empty regular objects
	if err := p.emptyBucket(ctx, client, bucket); err != nil {
		return types.DestroyResourceResult{
			Kind:   "S3 Bucket",
			Name:   bucket,
			Status: types.DestroyStatusFailed,
			Error:  fmt.Sprintf("failed to empty bucket: %v", err),
		}
	}

	_, err = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: &bucket,
	})
	if err != nil {
		contextMsg := "failed to delete bucket"
		switch {
		case strings.Contains(err.Error(), "AccessDenied"):
			contextMsg = "permission denied while deleting bucket"
		case strings.Contains(err.Error(), "BucketNotEmpty"):
			contextMsg = "bucket not empty during delete operation"
		case strings.Contains(err.Error(), "Conflict"):
			contextMsg = "bucket has pending operations preventing deletion"
		}

		return types.DestroyResourceResult{
			Kind:   "S3 Bucket",
			Name:   bucket,
			Status: types.DestroyStatusFailed,
			Error:  fmt.Sprintf("%s: %v", contextMsg, err),
		}
	}

	return types.DestroyResourceResult{
		Kind:   "S3 Bucket",
		Name:   bucket,
		Status: types.DestroyStatusDeleted,
		Error:  "",
	}
}

func (p *AwsCurDestroyPlugin) destroyWorkgroup(
	ctx context.Context,
	client *athena.Client,
	wg string,
) types.DestroyResourceResult {
	// Check if workgroup exists
	_, err := client.GetWorkGroup(ctx, &athena.GetWorkGroupInput{
		WorkGroup: &wg,
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "ResourceNotFoundException", "InvalidRequestException":
				return types.DestroyResourceResult{
					Kind:   "Athena Workgroup",
					Name:   wg,
					Status: types.DestroyStatusSkipped,
					Error:  "Workgroup already deleted or does not exist",
				}
			default:
				return types.DestroyResourceResult{
					Kind:   "Athena Workgroup",
					Name:   wg,
					Status: types.DestroyStatusFailed,
					Error:  fmt.Sprintf("failed to describe workgroup (%s): %v", apiErr.ErrorCode(), err),
				}
			}
		}

		return types.DestroyResourceResult{
			Kind:   "Athena Workgroup",
			Name:   wg,
			Status: types.DestroyStatusFailed,
			Error:  fmt.Sprintf("failed to describe workgroup: %v", err),
		}
	}

	// Attempt to delete
	_, err = client.DeleteWorkGroup(ctx, &athena.DeleteWorkGroupInput{
		WorkGroup:             &wg,
		RecursiveDeleteOption: aws.Bool(true),
	})
	if err != nil {
		return types.DestroyResourceResult{
			Kind:   "Athena Workgroup",
			Name:   wg,
			Status: types.DestroyStatusFailed,
			Error:  err.Error(),
		}
	}

	// Success
	return types.DestroyResourceResult{
		Kind:   "Athena Workgroup",
		Name:   wg,
		Status: types.DestroyStatusDeleted,
		Error:  "",
	}
}

func (p *AwsCurDestroyPlugin) isBucketEmpty(
	ctx context.Context,
	client *s3.Client,
	bucket string,
) (bool, error) {
	out, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  &bucket,
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		// If bucket does NOT exist → treat as empty
		if strings.Contains(err.Error(), "NoSuchBucket") ||
			strings.Contains(err.Error(), "NotFound") ||
			strings.Contains(err.Error(), "404") {
			return true, nil
		}
		return false, err
	}
	if len(out.Contents) > 0 {
		return false, nil
	}

	// Also check versions (in case versioning was enabled)
	ver, err := client.ListObjectVersions(ctx, &s3.ListObjectVersionsInput{
		Bucket:  &bucket,
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return false, err
	}

	if len(ver.Versions) > 0 || len(ver.DeleteMarkers) > 0 {
		return false, nil
	}
	return true, nil
}

func (p *AwsCurDestroyPlugin) emptyBucket(ctx context.Context, client *s3.Client, bucket string) error {
	pager := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{Bucket: &bucket})

	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}
		if len(page.Contents) == 0 {
			continue
		}

		var ids []s3types.ObjectIdentifier
		for _, obj := range page.Contents {
			ids = append(ids, s3types.ObjectIdentifier{Key: obj.Key})
		}

		_, err = client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: &bucket,
			Delete: &s3types.Delete{Objects: ids},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *AwsCurDestroyPlugin) deleteObjectVersions(ctx context.Context, client *s3.Client, bucket string) error {
	pager := s3.NewListObjectVersionsPaginator(client, &s3.ListObjectVersionsInput{
		Bucket: &bucket,
	})

	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}

		var ids []s3types.ObjectIdentifier
		for _, v := range page.Versions {
			ids = append(ids, s3types.ObjectIdentifier{Key: v.Key, VersionId: v.VersionId})
		}
		for _, m := range page.DeleteMarkers {
			ids = append(ids, s3types.ObjectIdentifier{Key: m.Key, VersionId: m.VersionId})
		}

		if len(ids) == 0 {
			continue
		}

		_, err = client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: &bucket,
			Delete: &s3types.Delete{Objects: ids},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func humanizePreviewError(err error, resource string) string {
	if err == nil {
		return ""
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchBucket":
			return "resource not found"
		case "ResourceNotFoundException":
			return "resource not found"
		case "AccessDenied", "AccessDeniedException":
			return "access denied"
		case "ThrottlingException", "TooManyRequestsException":
			return "request throttled"
		case "InvalidRequestException":
			return "invalid request"
		}
		if code := apiErr.ErrorCode(); code != "" {
			return code
		}
	}

	msg := err.Error()
	// Remove the resource name to shorten the message if present
	if resource != "" {
		msg = strings.ReplaceAll(msg, resource, "[resource]")
	}
	if len(msg) > 140 {
		return fmt.Sprintf("%s…", msg[:137])
	}
	return msg
}
