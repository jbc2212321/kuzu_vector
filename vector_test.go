package main

import (
	"context"
	"fmt"
	"testing"
)

func TestVector(t *testing.T) {
	vector := &VectorDB{}
	vector.InitConn(true).LoadVector().LoadVectorFunc().CreateVectorNode().CreateVectorIndex().CreateVectorExampleData()
}

func TestAddVectorIndex(t *testing.T) {

	vector := &VectorDB{}
	vector.InitConn(false).LoadVector().CreateVectorIndex()

}

func TestMatch(t *testing.T) {

	vector := &VectorDB{}
	vector.InitConn(false)

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
	vector.InitConn(false).LoadVector().LoadVectorFunc()

	query_vector, _ := vector.vectorFunc(context.Background(), "quantum machine learning")

	query := `
		CALL QUERY_VECTOR_INDEX(
			'Book',
			'title_vec_index',
			$query_vector,
			2
		)
		RETURN node.title ORDER BY distance;
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
