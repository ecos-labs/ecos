-- service_code = 'awskms'
{% macro aws_mappings_kms_category(usage_type_col='usage_type') %}
    case
        -- KMS key storage
        when {{ usage_type_col }} like '%KMS-Keys%' then 'KMS Keys [Management]'

        -- Standard KMS requests
        when {{ usage_type_col }} like '%KMS-Requests%' and {{ usage_type_col }} not like '%Asymmetric%' and {{ usage_type_col }} not like '%GenerateDatakeyPair%' then 'KMS Standard Requests [Security]'

        -- Asymmetric key requests (general)
        when {{ usage_type_col }} like '%Requests-Asymmetric%' and {{ usage_type_col }} not like '%RSA_2048%' then 'KMS Asymmetric Requests [Security]'

        -- RSA 2048 asymmetric requests (higher cost)
        when {{ usage_type_col }} like '%KMS-Requests-Asymmetric-RSA_2048%' then 'KMS RSA 2048 Requests [Security]'

        -- Data key pair generation requests
        when {{ usage_type_col }} like '%KMS-Requests-GenerateDatakeyPair-RSA%' then 'KMS RSA Data Key Pair Generation [Security]'
        when {{ usage_type_col }} like '%Requests-GenerateDatakeyPair-ECC%' then 'KMS ECC Data Key Pair Generation [Security]'

        else 'KMS [Other]'
    end
{% endmacro %}
