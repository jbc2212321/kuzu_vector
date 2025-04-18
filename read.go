package main

import (
	"context"
	"fmt"
	"strings"
)

type Read struct {
	vector *VectorDB
}

func (r *Read) OnStart() {
	r.vector = &VectorDB{}
	r.vector.InitConn(false, true)
	r.vector.LoadFTS()
	r.vector.LoadVector()
	r.vector.LoadVectorFunc() // 准备向量模型

}
func (r *Read) OnEnd() {
	defer func() {
		r.vector.OnEnd()
	}()
}

func (r *Read) QueryEntityForFTS(question string) {
	query :=
		`
			CALL QUERY_FTS_INDEX('Entity', 'Entity_fts_index', '%s', conjunctive := false)
			return node.entity_name,node.entity_description,score
			ORDER BY score DESC;
		`
	result, err := r.vector.conn.Query(fmt.Sprintf(query, question))
	if err != nil {
		return
	}
	for result.HasNext() {
		tuple, err := result.Next()
		if err != nil {
			panic(err)
		}
		defer tuple.Close()
		asMap, _ := tuple.GetAsMap()
		entityName := asMap["node.entity_name"].(string)
		entityDescription := asMap["node.entity_description"].(string)
		fmt.Println(entityName, entityDescription)
	}

}

func (r *Read) QueryRelationShipForFTS(question string) {
	query :=
		`
			CALL QUERY_FTS_INDEX('EntityRelationship', 'EntityRelationship_fts_index', '%s', conjunctive := false)
			return node.source_entity,node.target_entity,node.relationship_description,score
			ORDER BY score DESC LIMIT 3;
		`
	result, err := r.vector.conn.Query(fmt.Sprintf(query, question))
	if err != nil {
		return
	}
	for result.HasNext() {
		tuple, err := result.Next()
		if err != nil {
			panic(err)
		}
		defer tuple.Close()
		asMap, _ := tuple.GetAsMap()
		sourceEntity := asMap["node.source_entity"].(string)
		targetEntity := asMap["node.target_entity"].(string)
		relationshipDescription := asMap["node.relationship_description"].(string)
		fmt.Println(sourceEntity, targetEntity, relationshipDescription)

	}

}

func (r *Read) QueryEntityForVector(question string) {

	questionVector, _ := r.vector.vectorFunc(context.Background(), question)
	query :=
		`
		CALL QUERY_VECTOR_INDEX(
			'Entity',
			'Entity_vec_index',
			$question_vector,
			3
		)
		return node.entity_name, node.entity_description, distance ORDER BY distance desc
		`
	pre, err := r.vector.conn.Prepare(query)
	if err != nil {
		return
	}
	result, err := r.vector.conn.Execute(pre, map[string]any{
		"question_vector": FloatListToAnyList(questionVector),
	})
	if err != nil {
		return
	}

	for result.HasNext() {
		tuple, err := result.Next()
		if err != nil {
			panic(err)
		}
		defer tuple.Close()
		asMap, _ := tuple.GetAsMap()
		entityName := asMap["node.entity_name"].(string)
		entityDescription := asMap["node.entity_description"].(string)
		fmt.Println(entityName, entityDescription)

	}
}

func (r *Read) QueryRelationShipForVector(question string) {

	questionVector, _ := r.vector.vectorFunc(context.Background(), question)
	query :=
		`
		CALL QUERY_VECTOR_INDEX(
			'EntityRelationship',
			'EntityRelationship_vec_index',
			$question_vector,
			3
		)
		return node.source_entity, node.target_entity,node.relationship_description, distance ORDER BY distance desc
		`
	pre, err := r.vector.conn.Prepare(query)
	if err != nil {
		return
	}
	result, err := r.vector.conn.Execute(pre, map[string]any{
		"question_vector": FloatListToAnyList(questionVector),
	})
	if err != nil {
		return
	}

	for result.HasNext() {
		tuple, err := result.Next()
		if err != nil {
			panic(err)
		}
		defer tuple.Close()
		asMap, _ := tuple.GetAsMap()
		sourceEntity := asMap["node.source_entity"].(string)
		targetEntity := asMap["node.target_entity"].(string)
		relationshipDescription := asMap["node.relationship_description"].(string)
		fmt.Println(sourceEntity, targetEntity, relationshipDescription)

	}
}

func (r *Read) QueryEntityByName(entityName ...string) {

	query := `
		match (e:Entity) where e.entity_name in %s return e.entity_name,e.entity_type,e.entity_description;
	`

	result, err := r.vector.conn.Query(fmt.Sprintf(query, arrToString(entityName...)))
	if err != nil {
		return
	}
	for result.HasNext() {
		tuple, err := result.Next()
		if err != nil {
			panic(err)
		}
		defer tuple.Close()
		asMap, _ := tuple.GetAsMap()
		fmt.Println(asMap)
	}
}

func arrToString(arr ...string) string {
	return "['" + strings.Join(arr, "','") + "']"
}
