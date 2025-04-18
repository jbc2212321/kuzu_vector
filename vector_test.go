package main

import (
	"context"
	"fmt"
	"testing"
)

func TestMatch(t *testing.T) {

	vector := &VectorDB{}
	vector.InitConn(false, true)

	//cypher := "MATCH (b:Book) WHERE b.title='The Quantum World' RETURN b"
	cypher := "CALL SHOW_INDEXES() RETURN *"
	query, err := vector.conn.Query(cypher)
	if err != nil {
		return
	}
	for query.HasNext() {
		tuple, err := query.Next()
		if err != nil {
			panic(err)
		}
		defer tuple.Close()
		// The tuple can also be converted to a string.
		fmt.Print(tuple.GetAsString())
	}
}

func TestVectorSearch(t *testing.T) {

	vector := &VectorDB{}
	vector.InitConn(false, true).LoadVector().LoadVectorFunc()

	query_vector, _ := vector.vectorFunc(context.Background(), "mcp是什么东西?")

	query := `
		CALL QUERY_VECTOR_INDEX(
			'Entity',
			'entity_vec_index',
			$query_vector,
			2
		)
		RETURN node.entity_name,node.entity_description ,distance ORDER BY distance;
	`
	prepare, err := vector.conn.Prepare(query)
	if err != nil {
		panic(err)
	}

	res, err := vector.conn.Execute(prepare, map[string]any{
		"query_vector": FloatListToAnyList(query_vector),
	})
	if err != nil {
		panic(err)
	}

	for res.HasNext() {
		tuple, err := res.Next()
		if err != nil {
			panic(err)
		}
		defer tuple.Close()
		// The tuple can also be converted to a string.
		fmt.Print(tuple.GetAsString())
	}

}

func TestFTSSearch(t *testing.T) {

	vector := &VectorDB{}
	vector.InitConn(false, true).LoadFTS()

	query := `
			CALL QUERY_FTS_INDEX('Entity', 'entity_fts_index', 'mcp', conjunctive := false)
			return node.entity_name,node.entity_description,score
			ORDER BY score DESC;
	`

	res, err := vector.conn.Query(query)
	if err != nil {
		panic(err)
	}

	for res.HasNext() {
		tuple, err := res.Next()
		if err != nil {
			panic(err)
		}
		defer tuple.Close()
		// The tuple can also be converted to a string.
		fmt.Print(tuple.GetAsString())
	}

}
