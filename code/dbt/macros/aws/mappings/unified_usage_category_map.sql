{% macro aws_mappings_unified_usage_category_map(service_code_col='service_code', operation_col='operation', usage_type_col='usage_type', instance_type_family_col='instance_type_family', engine_col='engine', billing_entity_col='billing_entity', service_name_col='service_name', purchase_option_col='purchase_option') %}
case

        -- marketplace
        when {{ billing_entity_col }} = 'AWS Marketplace' then {{ service_name_col }}

        -- elastic load balancing
        when {{ service_code_col }} = 'AWSELB'
            then {{ aws_mappings_elb_category(usage_type_col=usage_type_col, operation_col=operation_col) }}

        -- networking (VPC, Direct Connect)
        when {{ service_code_col }} in ('AmazonVPC', 'AWSDirectConnect')
            then {{ aws_mappings_network_category(service_code_col=service_code_col, usage_type_col=usage_type_col, operation_col=operation_col) }}

        -- ec2
        when {{ service_code_col }} = 'AmazonEC2'
            then {{ aws_mappings_ec2_category(service_code_col=service_code_col, usage_type_col=usage_type_col, instance_type_family_col=instance_type_family_col, purchase_option_col=purchase_option_col) }}

        -- s3
        when {{ service_code_col }} = 'AmazonS3'
            then {{ aws_mappings_s3_category(operation_col=operation_col, usage_type_col=usage_type_col) }}

        -- elasticache
        when {{ service_code_col }} = 'AmazonElastiCache'
            then {{ aws_mappings_elasticache_category(usage_type_col='usage_type') }}

        -- opensearch
        when {{ service_code_col }} = 'AmazonES'
            then {{ aws_mappings_opensearch_category(usage_type_col=usage_type_col) }}

        -- rds
        when {{ service_code_col }} = 'AmazonRDS'
            then {{ aws_mappings_rds_category(usage_type_col=usage_type_col, engine_col=engine_col, operation_col=operation_col) }}

        -- dynamodb
        when {{ service_code_col }} = 'AmazonDynamoDB'
            then {{ aws_mappings_dynamodb_category(usage_type_col=usage_type_col) }}

        -- redshift
        when {{ service_code_col }} = 'AmazonRedshift'
            then {{ aws_mappings_redshift_category(usage_type_col=usage_type_col) }}

        -- documentdb
        when {{ service_code_col }} = 'AmazonDocDB'
            then {{ aws_mappings_documentdb_category(usage_type_col=usage_type_col) }}

        -- memorydb
        when {{ service_code_col }} = 'AmazonMemoryDB'
            then {{ aws_mappings_memorydb_category(usage_type_col=usage_type_col, engine_col=engine_col) }}

        -- neptune
        when {{ service_code_col }} = 'AmazonNeptune'
            then {{ aws_mappings_neptune_category(usage_type_col=usage_type_col) }}

        -- ecs
        when {{ service_code_col }} = 'AmazonECS'
            then {{ aws_mappings_ecs_category(usage_type_col=usage_type_col) }}

        -- eks
        when {{ service_code_col }} = 'AmazonEKS'
            then {{ aws_mappings_eks_category(usage_type_col=usage_type_col) }}

        -- cloudwatch
        when {{ service_code_col }} = 'AmazonCloudWatch'
            then {{ aws_mappings_cloudwatch_category(usage_type_col=usage_type_col) }}

        -- quicksight
        when {{ service_code_col }} = 'AmazonQuickSight'
            then {{ aws_mappings_quicksight_category(usage_type_col=usage_type_col) }}

        -- lambda
        when {{ service_code_col }} = 'AWSLambda'
            then {{ aws_mappings_lambda_category(usage_type_col=usage_type_col) }}

        -- batch
        when {{ service_code_col }} = 'AWSBatch'
            then {{ aws_mappings_batch_category(service_code_col=service_code_col, usage_type_col=usage_type_col) }}

        -- athena
        when {{ service_code_col }} = 'AmazonAthena'
            then {{ aws_mappings_athena_category(usage_type_col=usage_type_col) }}

        -- efs
        when {{ service_code_col }} = 'AmazonEFS'
            then {{ aws_mappings_efs_category(usage_type_col=usage_type_col) }}

        -- guardduty
        when {{ service_code_col }} = 'AmazonGuardDuty'
            then {{ aws_mappings_guardduty_category(usage_type_col=usage_type_col) }}

        -- kinesis
        when {{ service_code_col }} = 'AmazonKinesis'
            then {{ aws_mappings_kinesis_category(usage_type_col=usage_type_col) }}

        -- msk
        when {{ service_code_col }} = 'AmazonMSK'
            then {{ aws_mappings_msk_category(usage_type_col=usage_type_col) }}

        -- sagemaker
        when {{ service_code_col }} = 'AmazonSageMaker'
            then {{ aws_mappings_sagemaker_category(usage_type_col=usage_type_col) }}

        -- backup
        when {{ service_code_col }} = 'AWSBackup'
            then {{ aws_mappings_backup_category(usage_type_col=usage_type_col) }}

        -- glue
        when {{ service_code_col }} = 'AWSGlue'
            then {{ aws_mappings_glue_category(usage_type_col=usage_type_col) }}

        -- emr
        when {{ service_code_col }} = 'ElasticMapReduce'
            then {{ aws_mappings_emr_category(usage_type_col=usage_type_col) }}

        -- kms
        when {{ service_code_col }} = 'awskms'
            then {{ aws_mappings_kms_category(usage_type_col=usage_type_col) }}

        -- secrets manager
        when {{ service_code_col }} = 'AWSSecretsManager'
            then {{ aws_mappings_secrets_manager_category(usage_type_col=usage_type_col) }}

        -- security hub
        when {{ service_code_col }} = 'AWSSecurityHub'
            then {{ aws_mappings_security_hub_category(usage_type_col=usage_type_col) }}

        -- cloudtrail
        when {{ service_code_col }} = 'AWSCloudTrail'
            then {{ aws_mappings_cloudtrail_category(usage_type_col=usage_type_col) }}

        -- config
        when {{ service_code_col }} = 'AWSConfig'
            then {{ aws_mappings_config_category(usage_type_col=usage_type_col) }}

        -- compute savings plans
        when {{ service_code_col }} = 'ComputeSavingsPlans'
            then {{ aws_mappings_compute_savings_plans_category(usage_type_col=usage_type_col) }}

        -- aws support enterprise
        when {{ service_code_col }} = 'AWSSupportEnterprise'
            then {{ aws_mappings_awssupport_category(usage_type_col=usage_type_col) }}

        -- aws waf
        when {{ service_code_col }} = 'awswaf'
            then {{ aws_mappings_awswaf_category(usage_type_col=usage_type_col) }}

        -- sqs
        when {{ service_code_col }} = 'AWSQueueService'
            then {{ aws_mappings_sqs_category(usage_type_col=usage_type_col) }}

        -- ecr
        when {{ service_code_col }} = 'AmazonECR'
            then {{ aws_mappings_ecr_category(usage_type_col=usage_type_col) }}

        -- kinesis analytics
        when {{ service_code_col }} = 'AmazonKinesisAnalytics'
            then {{ aws_mappings_kinesis_analytics_category(usage_type_col=usage_type_col) }}

        -- kinesis firehose
        when {{ service_code_col }} = 'AmazonKinesisFirehose'
            then {{ aws_mappings_kinesis_firehose_category(usage_type_col=usage_type_col) }}

        -- sns
        when {{ service_code_col }} = 'AmazonSNS'
            then {{ aws_mappings_sns_category(usage_type_col=usage_type_col) }}

        -- bedrock
        when {{ service_code_col }} = 'AmazonBedrock'
            then {{ aws_mappings_bedrock_category(usage_type_col=usage_type_col) }}

        -- cognito
        when {{ service_code_col }} = 'AmazonCognito'
            then {{ aws_mappings_cognito_category(usage_type_col=usage_type_col) }}

        -- shield
        when {{ service_code_col }} = 'AWSShield'
            then {{ aws_mappings_shield_category(usage_type_col=usage_type_col) }}

        -- api gateway
        when {{ service_code_col }} = 'AmazonApiGateway'
            then {{ aws_mappings_api_gateway_category(usage_type_col=usage_type_col, operation_col=operation_col) }}

        -- cloudfront
        when {{ service_code_col }} = 'AmazonCloudFront'
            then {{ aws_mappings_cloudfront_category(usage_type_col=usage_type_col, operation_col=operation_col) }}

        -- route53
        when {{ service_code_col }} = 'AmazonRoute53'
            then {{ aws_mappings_route53_category(usage_type_col=usage_type_col, operation_col=operation_col) }}

        -- newly added services
        -- cost explorer
        when {{ service_code_col }} = 'AWSCostExplorer'
            then {{ aws_mappings_cost_explorer_category(usage_type_col=usage_type_col) }}

        -- cloudhsm
        when {{ service_code_col }} = 'CloudHSM'
            then {{ aws_mappings_cloudhsm_category(usage_type_col=usage_type_col) }}

        -- transfer
        when {{ service_code_col }} = 'AWSTransfer'
            then {{ aws_mappings_transfer_category(usage_type_col=usage_type_col) }}

        -- x-ray
        when {{ service_code_col }} = 'AWSXRay'
            then {{ aws_mappings_xray_category(usage_type_col=usage_type_col) }}

        -- certificate manager
        when {{ service_code_col }} = 'AWSCertificateManager'
            then {{ aws_mappings_certificate_manager_category(usage_type_col=usage_type_col) }}

        -- outposts
        when {{ service_code_col }} = 'AmazonOutposts'
            then {{ aws_mappings_outposts_category(service_code_col=service_code_col, usage_type_col=usage_type_col) }}

        -- mq
        when {{ service_code_col }} = 'AmazonMQ'
            then {{ aws_mappings_mq_category(usage_type_col=usage_type_col) }}

        -- dax
        when {{ service_code_col }} = 'AmazonDAX'
            then {{ aws_mappings_dax_category(usage_type_col=usage_type_col) }}

        else 'Others'
end

{% endmacro %}
