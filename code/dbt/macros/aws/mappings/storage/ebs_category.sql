-- service_code = 'AmazonEC2'
{% macro aws_mappings_ebs_category(service_code_col='service_code', usage_type_col='usage_type') %}
    case
        when {{ usage_type_col }} like '%EBS:%'
            then
                case split_part({{ usage_type_col }}, ':', 2)

                    -- provisioned storage (GB-month)
                    when 'VolumeUsage.gp3' then 'EBS Volume gp3 [Storage]'
                    when 'VolumeUsage.gp2' then 'EBS Volume gp2 [Storage]'
                    when 'VolumeUsage.io2' then 'EBS Volume io2 [Storage]'
                    when 'VolumeUsage.piops' then 'EBS Volume io1 [Storage]'
                    when 'VolumeUsage.st1' then 'EBS Volume st1 [Storage]'
                    when 'VolumeUsage.sc1' then 'EBS Volume sc1 [Storage]'
                    when 'VolumeUsage' then 'EBS Volume magnetic [Storage]'

                    -- provisioned IOPS/IO
                    when 'VolumeP-IOPS.gp3' then 'EBS Volume gp3 [IOPS]'
                    when 'VolumeP-IOPS.io2' then 'EBS Volume io2 [IOPS]'
                    when 'VolumeP-IOPS.io2.tier2' then 'EBS Volume io2 [IOPS]'
                    when 'VolumeP-IOPS.io2.tier3' then 'EBS Volume io2 [IOPS]'
                    when 'VolumeP-IOPS.piops' then 'EBS Volume io1 [IOPS]'
                    when 'VolumeIOUsage' then 'EBS Volume magnetic [IO]'

                    -- provisioned throughput
                    when 'VolumeP-Throughput.gp3' then 'EBS Volume gp3 [Throughput]'

                    -- snapshot storage
                    when 'SnapshotUsage' then 'EBS Volume snapshot [Storage]'
                    when 'SnapshotArchiveStorage' then 'EBS Volume snapshot archive [Storage]'

                    -- snapshot restore and api
                    when 'SnapshotArchiveEarlyDelete' then 'EBS Volume snapshot archive [Early Delete]'
                    when 'SnapshotArchiveRetrieval' then 'EBS Volume snapshot archive [Retrieval]'
                    when 'FastSnapshotRestore' then 'EBS Volume snapshot [Fast Restore]'
                    when 'directAPI.snapshot.List' then 'EBS Volume snapshot [Direct API]'
                    when 'directAPI.snapshot.Get' then 'EBS Volume snapshot [Direct API]'
                    when 'directAPI.snapshot.Put' then 'EBS Volume snapshot [Direct API]'

                    else 'EBS [Other]'
                end
    end
{% endmacro %}
