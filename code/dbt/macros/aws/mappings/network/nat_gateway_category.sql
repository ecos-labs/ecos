{% macro aws_mappings_nat_gateway_category(usage_type_col='usage_type') %}
    case
        -- NAT Gateway Data Processing
        when {{ usage_type_col }} like '%NatGateway-Bytes' then 'NAT Gateway Data Processing Charge'

        -- NAT Gateway Hourly Charges
        when {{ usage_type_col }} like '%NatGateway-Hours' then 'NAT Gateway Hourly Charge'

        -- Data Transfer In
        when {{ usage_type_col }} like '%In-Bytes' then 'NAT Gateway Data Transfer In'

        -- Data Transfer Out
        when {{ usage_type_col }} like '%Out-Bytes' then 'NAT Gateway Data Transfer Out'

        -- Regional Data Transfer
        when {{ usage_type_col }} like '%Regional-Bytes' then 'NAT Gateway Data Transfer Same Region'

        else {{ usage_type_col }}
    end
{% endmacro %}
