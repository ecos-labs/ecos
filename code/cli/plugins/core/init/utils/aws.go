package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	v1endpoints "github.com/aws/aws-sdk-go/aws/endpoints"
)

// ================= AWS utility functions =================

// ValidateAWSCredentials validates AWS credentials with timeout
func ValidateAWSCredentials(ctx context.Context, timeout time.Duration) error {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	_, err = stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("failed to validate AWS credentials: %w", err)
	}

	return nil
}

// ValidateGlueTable validates that a Glue table exists
func ValidateGlueTable(ctx context.Context, catalog, database, table, profile string) error {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config with profile %s: %w", profile, err)
	}

	glueClient := glue.NewFromConfig(cfg)
	_, err = glueClient.GetTable(ctx, &glue.GetTableInput{
		CatalogId:    aws.String(catalog),
		DatabaseName: aws.String(database),
		Name:         aws.String(table),
	})
	if err != nil {
		return fmt.Errorf("failed to get Glue table %s.%s.%s: %w", catalog, database, table, err)
	}

	return nil
}

// GetAWSAccountAndRegion retrieves the current AWS account ID and region
func GetAWSAccountAndRegion(ctx context.Context, timeout time.Duration) (accountID, region string, err error) {
	return GetAWSAccountAndRegionWithProfile(ctx, timeout, "")
}

// GetAWSAccountAndRegionWithProfile retrieves the current AWS account ID and region using a specific profile
func GetAWSAccountAndRegionWithProfile(ctx context.Context, timeout time.Duration, profile string) (accountID, region string, err error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	var cfg aws.Config
	if profile != "" && profile != "default" {
		cfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithSharedConfigProfile(profile))
	} else {
		cfg, err = awsconfig.LoadDefaultConfig(ctx)
	}
	if err != nil {
		return "", "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	result, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", "", fmt.Errorf("failed to get caller identity: %w", err)
	}

	return *result.Account, cfg.Region, nil
}

// IsValidRegion checks if the provided region is a valid AWS region (offline, using v1 endpoints)
func IsValidRegion(region string) bool {
	resolver := v1endpoints.DefaultResolver()
	partitions := resolver.(v1endpoints.EnumPartitions).Partitions()

	for _, p := range partitions {
		for id := range p.Regions() {
			if region == id {
				return true
			}
		}
	}
	return false
}
