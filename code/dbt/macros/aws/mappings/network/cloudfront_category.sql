-- service_code = 'AmazonCloudFront'
{% macro aws_mappings_cloudfront_category(usage_type_col='usage_type', operation_col='operation') %}
    case
        -- Requests
        when {{ usage_type_col }} like '%Requests%' then 'CloudFront [Requests]'

        -- Data transfer
        when {{ usage_type_col }} like '%Out-Bytes%' then 'CloudFront [Data Transfer]'
        when {{ usage_type_col }} like '%DataTransfer%' then 'CloudFront [Data Transfer]'

        -- Origin Shield
        when {{ usage_type_col }} like '%OriginShield%' then 'CloudFront [Origin Shield]'
        when {{ usage_type_col }} like '%Origin%' then 'CloudFront [Origin]'

        -- Lambda@Edge and CloudFront Functions
        when {{ usage_type_col }} like '%Lambda%' then 'CloudFront [Edge Compute]'
        when {{ usage_type_col }} like '%Function%' then 'CloudFront [Edge Compute]'

        -- Field Level Encryption
        when {{ usage_type_col }} like '%Encrypt%' then 'CloudFront [Encryption]'

        -- Real-time logs
        when {{ usage_type_col }} like '%Log%' then 'CloudFront [Logs]'

        -- Invalidation
        when {{ usage_type_col }} like '%Invalidation%' then 'CloudFront [Invalidation]'

        -- SSL certificates
        when {{ usage_type_col }} like '%SSL%' then 'CloudFront [SSL]'
        when {{ usage_type_col }} like '%DedicatedIP%' then 'CloudFront [SSL]'

        -- Default fallback
        else 'CloudFront [Other]'
    end
{% endmacro %}
