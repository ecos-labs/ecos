-- service_code = 'AmazonECR'
{% macro aws_mappings_ecr_category(usage_type_col='usage_type') %}
    case
        -- Storage
        when {{ usage_type_col }} like '%TimedStorage-ByteHrs%' then 'ECR Container Storage [Storage]'

        -- Data transfer patterns
        when {{ usage_type_col }} like '%AWS-Out-Bytes%' then 'ECR Inter-Region Transfer [Network]'
        when {{ usage_type_col }} like '%DataTransfer-Out-Bytes%' then 'ECR Data Transfer Out [Network]'
        when {{ usage_type_col }} like '%DataXfer-Out%' then 'ECR Data Transfer Out [Network]'

        else 'ECR [Other]'
    end
{% endmacro %}
