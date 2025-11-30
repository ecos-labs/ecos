{% macro aws_mappings_s3_bucket_keywords(bucket_name_col='resource_id') %}
    case
        -- Data storage and processing
        when regexp_like(lower({{ bucket_name_col }}), '\b(athena)\b') and regexp_like(lower({{ bucket_name_col }}), '\b(results|query)\b') then 'athena_results'
        when regexp_like(lower({{ bucket_name_col }}), '\b(backup|bak|bkp|snapshot|snap)\b') then 'backup'
        when regexp_like(lower({{ bucket_name_col }}), '\b(datalake|data-lake|lake)\b') then 'datalake'
        when regexp_like(lower({{ bucket_name_col }}), '\b(source|raw|ingest|landing)\b') then 'source'
        when regexp_like(lower({{ bucket_name_col }}), '\b(archive|cold|glacier|ia)\b') then 'archive'
        when regexp_like(lower({{ bucket_name_col }}), '\b(analytics|data|warehouse|dwh|etl|pipeline)\b') then 'analytics'
        when regexp_like(lower({{ bucket_name_col }}), '\b(processed|curated|prepared|transformed)\b') then 'processed'

        -- Content and media
        when regexp_like(lower({{ bucket_name_col }}), '\b(media|image|video|photo|asset|content)\b') then 'media'
        when regexp_like(lower({{ bucket_name_col }}), '\b(document|doc|file|pdf|office)\b') then 'document'
        when regexp_like(lower({{ bucket_name_col }}), '\b(web|www|static|frontend|ui)\b') then 'web'

        -- Infrastructure and operations
        when regexp_like(lower({{ bucket_name_col }}), '\b(log|audit|trail|event)\b') then 'logging'
        when regexp_like(lower({{ bucket_name_col }}), '\b(temp|tmp|staging|stage|test|dev|sandbox|qa)\b') then 'temp'
        when regexp_like(lower({{ bucket_name_col }}), '\b(config|setting|env|parameter|secret)\b') then 'config'

        -- Data flow and transfers
        when regexp_like(lower({{ bucket_name_col }}), '\b(upload|incoming|input)\b') then 'uploads'
        when regexp_like(lower({{ bucket_name_col }}), '\b(export|outgoing|output|extract)\b') then 'export'
        when regexp_like(lower({{ bucket_name_col }}), '\b(import|ingest)\b') then 'import'
        when regexp_like(lower({{ bucket_name_col }}), '\b(replication|replica|sync|mirror)\b') then 'replication'

        -- User and access patterns
        when regexp_like(lower({{ bucket_name_col }}), '\b(user|client|customer)\b') then 'user_data'
        when regexp_like(lower({{ bucket_name_col }}), '\b(public|shared|common|global)\b') then 'public'
        when regexp_like(lower({{ bucket_name_col }}), '\b(private|internal|corp|company)\b') then 'private'

        -- Specialized services
        when regexp_like(lower({{ bucket_name_col }}), '\b(ml|model|sagemaker|train|inference|ai)\b') then 'ml'
        when regexp_like(lower({{ bucket_name_col }}), '\b(report|dash|dashboard|bi|viz)\b') then 'reporting'
        when regexp_like(lower({{ bucket_name_col }}), '\b(cdn|edge|distribution)\b') then 'cdn'
        when regexp_like(lower({{ bucket_name_col }}), '\b(cache|redis|mem)\b') then 'cache'

        -- Legacy and specialized
        when regexp_like(lower({{ bucket_name_col }}), '\b(historical|history|legacy|old)\b') then 'historical'
        when regexp_like(lower({{ bucket_name_col }}), '\b(system|service|infra|infrastructure)\b') then 'system'
        when regexp_like(lower({{ bucket_name_col }}), '\b(compliance|legal|governance)\b') then 'compliance'
        when regexp_like(lower({{ bucket_name_col }}), '\b(security|sec|vault|encryption)\b') then 'security'

        else 'none'
    end
{% endmacro %}
