{% macro aws_transforms_usage_type_region_removal(usage_type_col='usage_type') %}
case
    -- NULL/empty checks first (fastest path, avoids regex evaluation)
    when {{ usage_type_col }} is null or {{ usage_type_col }} = '' then {{ usage_type_col }}

    -- Check for inter-region transfers first (13.88% of cases)
    -- Must be before single regions to prevent partial matches
    -- Examples: EUC1-USE1-AWS-Out-Bytes → AWS-Out-Bytes
    when regexp_like({{ usage_type_col }}, '^[A-Z]{2,4}[0-9]-[A-Z]{2,4}[0-9]-')
        then regexp_replace({{ usage_type_col }}, '^[A-Z]{2,4}[0-9]-[A-Z]{2,4}[0-9]-', '')

    -- Single region codes with digit (75.60% of cases - most common)
    -- Examples: USE2-BoxUsage → BoxUsage, EUC1-DataTransfer → DataTransfer
    when regexp_like({{ usage_type_col }}, '^[A-Z]{2,4}[0-9]-')
        then regexp_replace({{ usage_type_col }}, '^[A-Z]{2,4}[0-9]-', '')

    -- Two-letter region codes (4.05% of cases)
    -- Only specific known region prefixes to avoid removing service prefixes
    -- Examples: EU-DataTransfer → DataTransfer, AP-Requests → Requests
    when regexp_like({{ usage_type_col }}, '^(EU|AP|US|AU|IN|SA|ME|CA)-')
        then regexp_replace({{ usage_type_col }}, '^(EU|AP|US|AU|IN|SA|ME|CA)-', '')

    -- Wavelength zones (1.62% of cases)
    -- Examples: USE1WL1BNA1-CloudFront → CloudFront
    when regexp_like({{ usage_type_col }}, '^[A-Z]{2,4}[0-9]WL[0-9][A-Z0-9]{3,4}-')
        then regexp_replace({{ usage_type_col }}, '^[A-Z]{2,4}[0-9]WL[0-9][A-Z0-9]{3,4}-', '')

    -- Full region names in lowercase format (0.05% of cases)
    -- Examples: us-east-1-KMS-Requests → KMS-Requests
    --           eu-central-2a-API-Requests → API-Requests
    when regexp_like({{ usage_type_col }}, '^[a-z]{2}-[a-z]+-[0-9]+[a-z]?-')
        then regexp_replace({{ usage_type_col }}, '^[a-z]{2}-[a-z]+-[0-9]+[a-z]?-', '')

    -- Region codes with colon separator (0.07% of cases)
    -- Only removes region pattern (ending with digit), preserves service prefixes
    -- Removes: USE1:PI_API → PI_API, CAN1:PI_LTR → PI_LTR
    -- Preserves: RDS:Mirror-PIOPS, EBS:SnapshotUsage (no digit after letters)
    when regexp_like({{ usage_type_col }}, '^[A-Z]{2,4}[0-9]:')
        then regexp_replace({{ usage_type_col }}, '^[A-Z]{2,4}[0-9]:', '')

    -- Global prefix pattern (0.01% of cases, case-insensitive)
    -- Examples: Global-Requests → Requests, global-DataTransfer → DataTransfer
    when regexp_like({{ usage_type_col }}, '^[Gg]lobal-')
        then regexp_replace({{ usage_type_col }}, '^[Gg]lobal-', '')

    else {{ usage_type_col }}
end
{% endmacro %}
