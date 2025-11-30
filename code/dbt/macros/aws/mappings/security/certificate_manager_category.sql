-- service_code = 'AWSCertificateManager'
{% macro aws_mappings_certificate_manager_category(usage_type_col='usage_type') %}
    case
        -- Private CA management
        when {{ usage_type_col }} like '%PaidPrivateCA%' then 'Certificate Manager Private CA [Security]'
        when {{ usage_type_col }} like '%FreePrivateCA%' then 'Certificate Manager Free Private CA [Security]'

        -- Certificate issuance
        when {{ usage_type_col }} like '%PrivateCertificatesIssued%' then 'Certificate Manager Private Certificates [Security]'
        when {{ usage_type_col }} like '%ShortLivedCertificatesIssued%' then 'Certificate Manager Short-Lived Certificates [Security]'
        when {{ usage_type_col }} like '%ShortLivedCertificatePrivateCA%' then 'Certificate Manager Short-Lived Private CA [Security]'

        -- OCSP services
        when {{ usage_type_col }} like '%OCSPPerQuery%' then 'Certificate Manager OCSP Queries [Security]'
        when {{ usage_type_col }} like '%OCSPResponseHandling%' then 'Certificate Manager OCSP Response Handling [Security]'

        else 'Certificate Manager [Other]'
    end
{% endmacro %}
