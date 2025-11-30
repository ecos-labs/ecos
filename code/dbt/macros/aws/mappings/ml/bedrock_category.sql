-- service_code = 'AmazonBedrock'
{% macro aws_mappings_bedrock_category(usage_type_col='usage_type') %}
    case
        -- Text model tokens (input vs output)
        when {{ usage_type_col }} like '%-input-tokens' then 'Bedrock Text Model Input Tokens [AI/ML]'
        when {{ usage_type_col }} like '%-output-tokens' then 'Bedrock Text Model Output Tokens [AI/ML]'

        -- Image generation models
        when {{ usage_type_col }} like '%TitanImageGenerator%' and {{ usage_type_col }} like '%T2I%' then 'Bedrock Text-to-Image Generation [AI/ML]'
        when {{ usage_type_col }} like '%TitanImageGenerator%' and {{ usage_type_col }} like '%I2I%' then 'Bedrock Image-to-Image Generation [AI/ML]'

        -- Embedding models
        when {{ usage_type_col }} like '%TitanEmbeddings%' and {{ usage_type_col }} like '%Text%' then 'Bedrock Text Embeddings [AI/ML]'
        when {{ usage_type_col }} like '%TitanEmbeddings%' and {{ usage_type_col }} like '%Image%' then 'Bedrock Image Embeddings [AI/ML]'

        -- Provisioned throughput
        when {{ usage_type_col }} like '%ProvisionedThroughput%' and {{ usage_type_col }} like '%ModelUnits%' then 'Bedrock Provisioned Throughput [AI/ML]'

        -- Model customization and storage
        when {{ usage_type_col }} like '%Customization-Storage%' then 'Bedrock Model Customization Storage [AI/ML]'

        -- Guardrails
        when {{ usage_type_col }} like '%Guardrail-ContentPolicyUnitsConsumed%' then 'Bedrock Guardrails Content Policy [Security]'
        when {{ usage_type_col }} like '%Guardrail-SensitiveInformationPolicy%' and {{ usage_type_col }} like '%Paid%' then 'Bedrock Guardrails Sensitive Info (Paid) [Security]'
        when {{ usage_type_col }} like '%Guardrail-SensitiveInformationPolicy%' and {{ usage_type_col }} like '%Free%' then 'Bedrock Guardrails Sensitive Info (Free) [Security]'
        when {{ usage_type_col }} like '%Guardrail-TopicPolicyUnitsConsumed%' then 'Bedrock Guardrails Topic Policy [Security]'
        when {{ usage_type_col }} like '%Guardrail-WordPolicyUnitsConsumed%' then 'Bedrock Guardrails Word Policy [Security]'

        else 'Bedrock [Other]'
    end
{% endmacro %}
