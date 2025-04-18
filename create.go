package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Create struct {
	vector *VectorDB
	output *Output
}

var (
	getOutputOnce sync.Once
)

func (c *Create) OnStart(needRemove bool) {
	c.vector = &VectorDB{}
	c.vector.InitConn(needRemove, false)
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

// CreateFTSIndex fts索引可以作用多列
func (c *Create) CreateFTSIndex(tableName string) *Create {
	propList := "['entity_name','entity_description']"
	if tableName == "EntityRelationship" {
		propList = "['source_entity','target_entity','relationship_description']"
	}
	query := `
	CALL CREATE_FTS_INDEX(
		'%s',   
		'%s_fts_index',  
		 %s,   
		stemmer := 'porter',
		stopwords := 'stopwords.csv'
	)
	`

	_, err := c.vector.conn.Query(fmt.Sprintf(query, tableName, tableName, propList))
	if err != nil {
		panic(err)
	}

	return c
}

// CreateVectorIndex 向量索引只能作用于单列
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
