{% macro aws_mappings_operating_system(operation_col='operation', operating_system_col='operating_system') %}
case
    when split_part({{ operation_col }}, ':', 2) = '0002' then 'Windows'
    when split_part({{ operation_col }}, ':', 2) = '0004' then 'Linux with SQL Server Standard'
    when split_part({{ operation_col }}, ':', 2) = '0006' then 'Windows with SQL Server Standard'
    when split_part({{ operation_col }}, ':', 2) = '000g' then 'SUSE Linux'
    when split_part({{ operation_col }}, ':', 2) = '0010' then 'Red Hat Enterprise Linux'
    when split_part({{ operation_col }}, ':', 2) = '0014' then 'Red Hat Enterprise Linux with SQL Server Standard'
    when split_part({{ operation_col }}, ':', 2) = '0100' then 'SQL Server Enterprise'
    when split_part({{ operation_col }}, ':', 2) = '0102' then 'Windows with SQL Server Enterprise'
    when split_part({{ operation_col }}, ':', 2) = '0110' then 'Red Hat Enterprise Linux with SQL Server Enterprise'
    when split_part({{ operation_col }}, ':', 2) = '0200' then 'SQL Server Web'
    when split_part({{ operation_col }}, ':', 2) = '0202' then 'Windows with SQL Server Web'
    when split_part({{ operation_col }}, ':', 2) = '0210' then 'Red Hat Enterprise Linux with SQL Server Web'
    when split_part({{ operation_col }}, ':', 2) = '0800' then 'Windows BYOL'
    when split_part({{ operation_col }}, ':', 2) = '0g00' then 'Ubuntu Pro'
    when split_part({{ operation_col }}, ':', 2) = '00g0' then 'Red Hat BYOL Linux'
    when split_part({{ operation_col }}, ':', 2) = '1010' then 'Red Hat Enterprise Linux with HA'
    when split_part({{ operation_col }}, ':', 2) = '1014' then 'Red Hat Enterprise Linux with HA and SQL Server Standard'
    when split_part({{ operation_col }}, ':', 2) = '1110' then 'Red Hat Enterprise Linux with HA and SQL Server Enterprise'
    when substr(split_part({{ operation_col }}, ':', 2), 1, 2) = 'SV' then 'Linux/UNIX'
    when {{ operation_col }} = 'RunInstances' then 'Linux/UNIX'
    else {{ operating_system_col }}
end
{% endmacro %}
