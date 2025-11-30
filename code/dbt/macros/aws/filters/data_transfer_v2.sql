{% macro aws_filters_data_transfer_v2(subservice_code_col='subservice_code', usage_type_col='usage_type') %}

case
    when {{ subservice_code_col }} = 'AWSDataTransfer' then true
    when {{ usage_type_col }} like '%DataXfer%' then true
    when {{ usage_type_col }} like '%DataTransfer%' then true
    when {{ usage_type_col }} like '%IntraRegion%' then true
    when {{ usage_type_col }} like '%InterRegion%' then true
    when {{ usage_type_col }} like '%Out-Bytes%' then true
    when {{ usage_type_col }} like '%In-Bytes%' then true
    when {{ usage_type_col }} like '%CloudFront-Out-Bytes%' then true
    when {{ usage_type_col }} like '%CloudFront-In-Bytes%' then true
    when {{ usage_type_col }} like '%Data-Bytes-In%' then true
    when {{ usage_type_col }} like '%Data-Bytes-Out%' then true
    else false
end

{% endmacro %}
