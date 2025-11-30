-- service_code = 'AWSSecretsManager'
{% macro aws_mappings_secrets_manager_category(usage_type_col='usage_type') %}
    case
        -- Secrets storage
        when {{ usage_type_col }} like '%AWSSecretsManager-Secrets%' then 'Secrets Manager Secrets [Management]'
        when {{ usage_type_col }} like '%AWSSecretsManager-Secret%' then 'Secrets Manager Secrets [Management]'

        -- API requests
        when {{ usage_type_col }} like '%AWSSecretsManagerAPIRequest%' then 'Secrets Manager API Requests [Security]'
        when {{ usage_type_col }} like '%AWSSecretsManager-APIRequests%' then 'Secrets Manager API Requests [Security]'

        else 'Secrets Manager [Other]'
    end
{% endmacro %}
