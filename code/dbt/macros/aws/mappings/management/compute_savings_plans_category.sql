-- service_code = 'ComputeSavingsPlans'
{% macro aws_mappings_compute_savings_plans_category(usage_type_col='usage_type') %}
    case
        -- Compute Savings Plans - 1 year commitment
        when {{ usage_type_col }} like '%ComputeSP:1yrNoUpfront%' then 'Compute Savings Plans 1yr No Upfront [Compute]'
        when {{ usage_type_col }} like '%ComputeSP:1yrPartialUpfront%' then 'Compute Savings Plans 1yr Partial Upfront [Compute]'
        when {{ usage_type_col }} like '%ComputeSP:1yrAllUpfront%' then 'Compute Savings Plans 1yr All Upfront [Compute]'

        -- Compute Savings Plans - 3 year commitment
        when {{ usage_type_col }} like '%ComputeSP:3yrNoUpfront%' then 'Compute Savings Plans 3yr No Upfront [Compute]'
        when {{ usage_type_col }} like '%ComputeSP:3yrPartialUpfront%' then 'Compute Savings Plans 3yr Partial Upfront [Compute]'
        when {{ usage_type_col }} like '%ComputeSP:3yrAllUpfront%' then 'Compute Savings Plans 3yr All Upfront [Compute]'

        -- SageMaker Savings Plans - 1 year commitment
        when {{ usage_type_col }} like '%SageMakerSP:1yrNoUpfront%' then 'SageMaker Savings Plans 1yr No Upfront [ML Compute]'
        when {{ usage_type_col }} like '%SageMakerSP:1yrPartialUpfront%' then 'SageMaker Savings Plans 1yr Partial Upfront [ML Compute]'
        when {{ usage_type_col }} like '%SageMakerSP:1yrAllUpfront%' then 'SageMaker Savings Plans 1yr All Upfront [ML Compute]'

        -- SageMaker Savings Plans - 3 year commitment
        when {{ usage_type_col }} like '%SageMakerSP:3yrNoUpfront%' then 'SageMaker Savings Plans 3yr No Upfront [ML Compute]'
        when {{ usage_type_col }} like '%SageMakerSP:3yrPartialUpfront%' then 'SageMaker Savings Plans 3yr Partial Upfront [ML Compute]'
        when {{ usage_type_col }} like '%SageMakerSP:3yrAllUpfront%' then 'SageMaker Savings Plans 3yr All Upfront [ML Compute]'

        else 'Compute Savings Plans [Other]'
    end
{% endmacro %}
