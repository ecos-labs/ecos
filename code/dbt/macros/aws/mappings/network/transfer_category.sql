-- service_code = 'AWSTransfer'
{% macro aws_mappings_transfer_category(usage_type_col='usage_type') %}
    case
        -- SFTP endpoints
        when {{ usage_type_col }} like '%SFTP%' and {{ usage_type_col }} like '%EndpointHour%' then 'Transfer SFTP Endpoint Hours [Network]'
        when {{ usage_type_col }} like '%SFTP%' and {{ usage_type_col }} like '%Upload%' then 'Transfer SFTP Upload [Network]'
        when {{ usage_type_col }} like '%SFTP%' and {{ usage_type_col }} like '%Download%' then 'Transfer SFTP Download [Network]'

        -- FTPS endpoints
        when {{ usage_type_col }} like '%FTPS%' and {{ usage_type_col }} like '%EndpointHour%' then 'Transfer FTPS Endpoint Hours [Network]'
        when {{ usage_type_col }} like '%FTPS%' and {{ usage_type_col }} like '%Upload%' then 'Transfer FTPS Upload [Network]'
        when {{ usage_type_col }} like '%FTPS%' and {{ usage_type_col }} like '%Download%' then 'Transfer FTPS Download [Network]'

        -- FTP endpoints
        when {{ usage_type_col }} like '%FTP%' and {{ usage_type_col }} like '%EndpointHour%' and {{ usage_type_col }} not like '%SFTP%' and {{ usage_type_col }} not like '%FTPS%' then 'Transfer FTP Endpoint Hours [Network]'
        when {{ usage_type_col }} like '%FTP%' and {{ usage_type_col }} like '%Upload%' and {{ usage_type_col }} not like '%SFTP%' and {{ usage_type_col }} not like '%FTPS%' then 'Transfer FTP Upload [Network]'
        when {{ usage_type_col }} like '%FTP%' and {{ usage_type_col }} like '%Download%' and {{ usage_type_col }} not like '%SFTP%' and {{ usage_type_col }} not like '%FTPS%' then 'Transfer FTP Download [Network]'

        -- AS2 endpoints
        when {{ usage_type_col }} like '%AS2%' and {{ usage_type_col }} like '%EndpointHour%' then 'Transfer AS2 Endpoint Hours [Network]'
        when {{ usage_type_col }} like '%AS2%' and {{ usage_type_col }} like '%Message%' then 'Transfer AS2 Messages [Network]'

        -- Data transfer
        when {{ usage_type_col }} like '%DataTransfer%' then 'Transfer Data Transfer [Network]'

        else 'Transfer [Other]'
    end
{% endmacro %}
