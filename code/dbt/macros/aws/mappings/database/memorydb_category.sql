-- service_code = 'AmazonMemoryDB'
{% macro aws_mappings_memorydb_category(usage_type_col='usage_type', engine_col='engine') %}
    case

        -- Instance
        when {{ usage_type_col }} like '%NodeUsage%' then 'MemoryDB ' || {{ engine_col }} || ' Instance [Compute]'

        -- Data writes
        when {{ usage_type_col }} like '%DataWritten%' then 'MemoryDB ' || {{ engine_col }} || ' Write [Storage]'

        -- Snapshots
        when {{ usage_type_col }} like '%SnapshotUsage%' then 'MemoryDB ' || {{ engine_col }} || ' Snapshot [Storage]'

        else 'MemoryDB [Other]'
    end
{% endmacro %}
