{% macro aws_filters_data_transfer_v1(product_family_col='product_family', charge_type_col='charge_type') %}

not (
    {{ product_family_col }} in ('Data Transfer', 'DT-Data Transfer')
    and {{ charge_type_col }} in ('DiscountedUsage', 'Usage', 'SavingsPlanCoveredUsage')
)

{% endmacro %}
