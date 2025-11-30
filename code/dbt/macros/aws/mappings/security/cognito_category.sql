-- service_code = 'AmazonCognito'
{% macro aws_mappings_cognito_category(usage_type_col='usage_type') %}
    case
        -- User pool operations
        when {{ usage_type_col }} like '%UserPool%' and {{ usage_type_col }} like '%MAU%' then 'Cognito User Pool Monthly Active Users [Processing]'
        when {{ usage_type_col }} like '%UserPool%' and {{ usage_type_col }} like '%SMS%' then 'Cognito User Pool SMS [Processing]'
        when {{ usage_type_col }} like '%UserPool%' and {{ usage_type_col }} like '%Email%' then 'Cognito User Pool Email [Processing]'
        when {{ usage_type_col }} like '%UserPool%' and {{ usage_type_col }} not like '%MAU%' and {{ usage_type_col }} not like '%SMS%' and {{ usage_type_col }} not like '%Email%' then 'Cognito User Pool Operations [Processing]'

        -- Identity pool operations
        when {{ usage_type_col }} like '%IdentityPool%' then 'Cognito Identity Pool Operations [Processing]'

        -- Advanced security features
        when {{ usage_type_col }} like '%AdvancedSecurity%' then 'Cognito Advanced Security Features [Security]'

        -- SAML operations
        when {{ usage_type_col }} like '%SAML%' then 'Cognito SAML Operations [Processing]'

        else 'Cognito [Other]'
    end
{% endmacro %}
