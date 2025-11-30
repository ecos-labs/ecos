-- service_code = 'AmazonSageMaker'
{% macro aws_mappings_sagemaker_category(usage_type_col='usage_type') %}
    case
        -- Storage (all types)
        when {{ usage_type_col }} like '%VolumeUsage%' then 'SageMaker Storage [Storage]'

        -- Core ML training and processing
        when {{ usage_type_col }} like '%Train:%' or {{ usage_type_col }} like '%Processing:%'
             or {{ usage_type_col }} like '%Tsform:%' or {{ usage_type_col }} like '%TrSpt:%' then 'SageMaker Training & Processing [Compute]'

        -- Inference and hosting
        when {{ usage_type_col }} like '%Host:%' or {{ usage_type_col }} like '%ServerlessInf:%'
             or {{ usage_type_col }} like '%AsyncInf:%' or {{ usage_type_col }} like '%ProvisionedConcurrency:%' then 'SageMaker Inference [Compute]'

        -- Development environments
        when {{ usage_type_col }} like '%Notebk:%' or {{ usage_type_col }} like '%Studio:%' then 'SageMaker Development [Compute]'

        -- HyperPod clusters
        when {{ usage_type_col }} like '%Cluster:%' then 'SageMaker HyperPod [Compute]'

        -- Feature Store operations
        when {{ usage_type_col }} like '%FeatureStore:%' then 'SageMaker Feature Store [Management]'

        -- Managed services
        when {{ usage_type_col }} like '%Canvas:%' or {{ usage_type_col }} like '%MLflow:%'
             or {{ usage_type_col }} like '%Geospatial:%' or {{ usage_type_col }} like '%TensorBoard:%' then 'SageMaker Managed Services [Management]'


        -- Data transfer (all types)
        when {{ usage_type_col }} like '%DataTransfer-%' or {{ usage_type_col }} like '%Data-Bytes%' then 'SageMaker Data Transfer [Network]'

        else 'SageMaker [Other]'
    end
{% endmacro %}
