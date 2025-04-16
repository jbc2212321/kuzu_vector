package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Output struct {
	EntityList       []*Entity           `json:"entityList"`
	RelationshipList []*RelationshipList `json:"relationshipList"`
}

type Entity struct {
	EntityName        string `json:"entity_name"`
	EntityType        string `json:"entity_type"`
	EntityDescription string `json:"entity_description"`
}

type RelationshipList struct {
	SourceEntity            string `json:"source_entity"`
	TargetEntity            string `json:"target_entity"`
	RelationshipDescription string `json:"relationship_description"`
	RelationshipStrength    int    `json:"relationship_strength"`
}

type Create struct {
	vector *VectorDB
	output *Output
}

var (
	getOutputOnce sync.Once
)

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

func (c *Create) OnCreate(needRemove bool) {
	c.vector = &VectorDB{}
	c.vector.InitConn(needRemove)
	c.vector.LoadFTS()
	c.vector.LoadVector()
	c.vector.LoadVectorFunc() // 准备向量模型

}

func (c *Create) CHECKPOINT() {
	_, err := c.vector.conn.Query("CHECKPOINT")
	if err != nil {
		panic(err)
	}
}

func (c *Create) GetOutPut() *Output {

	getOutputOnce.Do(func() {
		data, err := os.ReadFile("output.json")
		if err != nil {
			panic(err)
		}

		var outputObj Output
		if err := json.Unmarshal(data, &outputObj); err != nil {
			panic(err)
		}
		c.output = &outputObj
	})
	return c.output
}

func (c *Create) OnEnd() {

	defer func() {
		c.vector.OnEnd()
	}()

}

func (c *Create) CreateNode() *Create {
	queries := []string{
		CreateEntityNode,
		CreateEntityRefNode,
		CreateEntityRefEdge,
	}
	for _, query := range queries {
		_, err := c.vector.conn.Query(query)
		if err != nil {
			panic(err)
		}
	}

	return c
}

func (c *Create) InsertNode() *Create {
	ctx := context.Background()
	for _, entity := range c.GetOutPut().EntityList {
		// 创建向量
		descriptionEmbedding, err := c.vector.vectorFunc(ctx, entity.EntityDescription)
		if err != nil {
			panic(err)
		}
		// 插入节点
		preparedStatement, err := c.vector.conn.Prepare(InsertEntityNode)
		if err != nil {
			return nil
		}
		_, err = c.vector.conn.Execute(preparedStatement,
			map[string]interface{}{
				"entity_name":           entity.EntityName,
				"entity_type":           entity.EntityType,
				"entity_description":    entity.EntityDescription,
				"description_embedding": FloatListToAnyList(descriptionEmbedding),
			},
		)
		if err != nil {
			panic(err)
		}
	}

	for _, relationship := range c.GetOutPut().RelationshipList {
		// 创建向量
		descriptionEmbedding, err := c.vector.vectorFunc(ctx, relationship.RelationshipDescription)
		if err != nil {
			panic(err)
		}
		// 插入节点
		preparedStatement, err := c.vector.conn.Prepare(InsertEntityRefNode)
		if err != nil {
			return nil
		}
		_, err = c.vector.conn.Execute(preparedStatement,
			map[string]interface{}{
				"source_entity":            relationship.SourceEntity,
				"target_entity":            relationship.TargetEntity,
				"relationship_description": relationship.RelationshipDescription,
				"description_embedding":    FloatListToAnyList(descriptionEmbedding),
			},
		)
		if err != nil {
			panic(err)
		}

		// 插入边
		preparedStatement, err = c.vector.conn.Prepare(InsertEntityRefEdge)
		if err != nil {
			return nil
		}
		_, err = c.vector.conn.Execute(preparedStatement,
			map[string]interface{}{
				"source_entity":            relationship.SourceEntity,
				"target_entity":            relationship.TargetEntity,
				"relationship_description": relationship.RelationshipDescription,
				"relationship_strength":    relationship.RelationshipStrength,
			},
		)
		if err != nil {
			panic(err)
		}

	}

	return c
}

func (c *Create) CreateFTSIndex(tableName string, property string) *Create {
	query := `
	CALL CREATE_FTS_INDEX(
		'%s',   
		'%s_fts_index',  
		['%s'],   
		stemmer := 'porter',
		stopwords := 'stopwords.csv'
	)
	`
	//

	_, err := c.vector.conn.Query(fmt.Sprintf(query, tableName, tableName, property))
	if err != nil {
		panic(err)
	}

	return c
}

func (c *Create) CreateVectorIndex(tableName string, property string) *Create {
	query := `
		CALL CREATE_VECTOR_INDEX('%s','%s_vec_index','%s')
	`

	_, err := c.vector.conn.Query(fmt.Sprintf(query, tableName, tableName, property))
	if err != nil {
		panic(err)
	}

	return c
}
