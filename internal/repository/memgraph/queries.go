package memgraph

const (
	QueryUpsertDataset = `
		MERGE (d:Dataset {id: $id})
		SET d.name = $name,
		    d.database = $database,
		    d.schema = $schema,
		    d.description = $description,
		    d.updated_at = timestamp()
		RETURN d
	`

	QueryUpsertColumn = `
		MERGE (c:Column {id: $id})
		SET c.name = $name,
		    c.data_type = $data_type,
		    c.nullable = $nullable,
		    c.dataset_id = $dataset_id,
		    c.ordinal_position = $ordinal_position
		RETURN c
	`

	QueryLinkDatasetColumn = `
		MATCH (d:Dataset {id: $dataset_id})
		MATCH (c:Column {id: $column_id})
		MERGE (d)-[r:HAS_COLUMN]->(c)
		SET r.ordinal_position = $ordinal_position
		RETURN r
	`

	QueryRecordLineage = `
		MATCH (src:Column {id: $source_id})
		MATCH (tgt:Column {id: $target_id})
		MERGE (tgt)-[r:DERIVED_FROM]->(src)
		SET r.transformation = $transformation,
		    r.job_id = $job_id,
		    r.timestamp = timestamp()
		RETURN r
	`

	QueryRecordDatasetLineage = `
		MATCH (src:Dataset {id: $source_id})
		MATCH (tgt:Dataset {id: $target_id})
		MERGE (src)-[r:TRANSFORMS_TO]->(tgt)
		SET r.job_id = $job_id,
		    r.timestamp = timestamp()
		RETURN r
	`

	QueryTraceUpstream = `
		MATCH path = (src)-[r:DERIVED_FROM*1..%d]->(tgt:Column {id: $column_id})
		WITH relationships(path) AS rels
		UNWIND rels AS r
		RETURN r.transformation AS transformation,
		       r.job_id AS job_id,
		       r.timestamp AS timestamp
		LIMIT 100
	`

	QueryTraceDownstream = `
		MATCH path = (src:Column {id: $column_id})-[r:DERIVED_FROM*1..%d]->(tgt)
		WITH relationships(path) AS rels
		UNWIND rels AS r
		RETURN r.transformation AS transformation,
		       r.job_id AS job_id,
		       r.timestamp AS timestamp
		LIMIT 100
	`

	QueryGetDatasetLineage = `
		MATCH (src)-[r:DERIVED_FROM|MAPPED_TO]->(tgt:Column)
		WHERE src.dataset_id = $dataset_id OR tgt.dataset_id = $dataset_id
		RETURN src.id AS source_id,
		       src.type AS source_type,
		       src.name AS source_name,
		       tgt.id AS target_id,
		       tgt.type AS target_type,
		       tgt.name AS target_name,
		       r.transformation AS transformation,
		       r.job_id AS job_id
	`

	QueryGetLineageGraph = `
		MATCH (n)-[r]->(m)
		WHERE n.dataset_id = $dataset_id OR m.dataset_id = $dataset_id
		RETURN n.id AS node_id,
		       n.type AS node_type,
		       n.name AS node_name,
		       m.id AS target_id,
		       m.type AS target_type,
		       m.name AS target_name,
		       type(r) AS relationship_type,
		       r.transformation AS transformation
	`

	QueryRecordJob = `
		MERGE (j:Job {id: $id})
		SET j.dag_id = $dag_id,
		    j.task_id = $task_id,
		    j.run_id = $run_id,
		    j.status = $status,
		    j.start_time = $start_time,
		    j.end_time = $end_time
		RETURN j
	`

	QueryLinkJobDataset = `
		MATCH (j:Job {id: $job_id})
		MATCH (d:Dataset {id: $dataset_id})
		MERGE (j)-[r:READS|WRITES]->(d)
		SET r.timestamp = timestamp()
		RETURN r
	`
)

var (
	QueryDeleteDataset = `
		MATCH (d:Dataset {id: $id})
		DETACH DELETE d
	`

	QueryDeleteColumn = `
		MATCH (c:Column {id: $id})
		DETACH DELETE c
	`

	QueryDeleteLineage = `
		MATCH ()-[r:DERIVED_FROM]->()
		WHERE r.timestamp < $before
		DELETE r
	`
)