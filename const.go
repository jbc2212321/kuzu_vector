package main

const (
	// CreateEntityNode 实体节点
	CreateEntityNode string = "CREATE NODE TABLE IF NOT EXISTS Entity(id SERIAL PRIMARY KEY,entity_name STRING,entity_type STRING, entity_description string, description_embedding FLOAT[1024])"
	// CreateEntityRefEdge 节点之间的关联关系
	CreateEntityRefEdge string = "CREATE REL TABLE IF NOT EXISTS Relationship(FROM Entity TO Entity, id SERIAL PRIMARY KEY, relationship_description string,relationship_strength INT)"
	// CreateEntityRefNode 关系节点，记录关系，只有节点可以创建向量索引、FTS索引
	CreateEntityRefNode string = "CREATE NODE TABLE IF NOT EXISTS EntityRelationship( id SERIAL PRIMARY KEY, source_entity STRING,target_entity STRING,relationship_description string,description_embedding FLOAT[1024] ,relationship_strength INT)"

	InsertEntityNode    = "CREATE (e:Entity { entity_name: $entity_name, entity_type:$entity_type,entity_description: $entity_description, description_embedding: $description_embedding})"
	InsertEntityRefNode = "CREATE (e:EntityRelationship { source_entity: $source_entity, target_entity:$target_entity,relationship_description: $relationship_description, description_embedding: $description_embedding})"
	InsertEntityRefEdge = "MATCH (e1:Entity {entity_name: $source_entity}) MATCH (e2:Entity {entity_name: $target_entity}) CREATE (e1)-[:Relationship{relationship_description:$relationship_description,relationship_strength:$relationship_strength}]->(e2);"
)
