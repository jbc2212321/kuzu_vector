package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kuzudb/go-kuzu"
	"github.com/philippgille/chromem-go"
)

type VectorDB struct {
	conn       *kuzu.Connection
	db         *kuzu.Database
	vectorFunc chromem.EmbeddingFunc
}

func (*VectorDB) Name() string {
	return "vector_db"
}

func (v *VectorDB) InitConn(needRemove bool) *VectorDB {
	dbPath := v.Name()
	if needRemove {
		// 创建新数据
		os.RemoveAll(dbPath)
	}

	systemConfig := kuzu.DefaultSystemConfig()
	systemConfig.BufferPoolSize = 1024 * 1024 * 1024
	db, err := kuzu.OpenDatabase(dbPath, systemConfig)
	if err != nil {
		panic(err)
	}

	conn, err := kuzu.OpenConnection(db)
	if err != nil {
		panic(err)
	}

	v.db = db
	v.conn = conn
	return v
}
func (v *VectorDB) LoadVectorFunc() *VectorDB {
	v.vectorFunc = EmbeddingFunc()
	return v
}

func (v *VectorDB) LoadVector() *VectorDB {
	installPre, err := v.conn.Prepare("INSTALL VECTOR;")
	if err != nil {
		panic(err)
	}
	_, err = v.conn.Execute(installPre, nil)
	if err != nil {
		panic(err)
	}

	loadPre, err := v.conn.Prepare("LOAD VECTOR;")
	if err != nil {
		panic(err)
	}
	_, err = v.conn.Execute(loadPre, nil)
	if err != nil {
		panic(err)
	}
	return v
}

func (v *VectorDB) LoadFTS() *VectorDB {
	installPre, err := v.conn.Prepare("INSTALL FTS;")
	if err != nil {
		panic(err)
	}
	_, err = v.conn.Execute(installPre, nil)
	if err != nil {
		panic(err)
	}

	loadPre, err := v.conn.Prepare("LOAD EXTENSION FTS;")
	if err != nil {
		panic(err)
	}
	_, err = v.conn.Execute(loadPre, nil)
	if err != nil {
		panic(err)
	}
	return v
}

func (v *VectorDB) OnEnd() {

	defer func() {
		v.db.Close()
		v.conn.Close()

	}()

}

func FloatListToAnyList(floats []float32) []any {
	var list []any
	for _, v := range floats {
		list = append(list, v)
	}
	return list
}

// 以下都是测试
func (v *VectorDB) CreateVectorNode() *VectorDB {
	queries := []string{
		"CREATE NODE TABLE Book(id SERIAL PRIMARY KEY , title STRING,abstract STRING, title_embedding FLOAT[1024], published_year INT64)",
		"CREATE NODE TABLE Publisher(name STRING PRIMARY KEY)",
		"CREATE REL TABLE PublishedBy(FROM Book TO Publisher)",
	}
	for _, query := range queries {
		queryResult, err := v.conn.Query(query)
		if err != nil {
			panic(err)
		}
		defer queryResult.Close()
	}

	return v
}

func (v *VectorDB) CreateFTSIndex() *VectorDB {
	query := `
	CALL CREATE_FTS_INDEX(
		'Book',   
		'book_index',  
		['title','abstract'],   
		stemmer := 'porter'
	)
	`
	//stopwords := 'stopwords.csv'

	_, err := v.conn.Query(query)
	if err != nil {
		panic(err)
	}

	return v
}

func (v *VectorDB) CreateVectorIndex() *VectorDB {
	index := "CALL CREATE_VECTOR_INDEX('Book','title_vec_index','title_embedding')"

	_, err := v.conn.Query(index)
	if err != nil {
		panic(err)
	}

	return v
}

func (v *VectorDB) CreateVectorExampleData() *VectorDB {
	titles := []string{
		"The Quantum World",
		"Chronicles of the Universe",
		"Learning Machines",
		"Echoes of the Past",
		"The Dragon's Call",
	}
	publishers := []string{"Harvard University Press", "Independent Publisher", "Pearson", "McGraw-Hill Ryerson", "O'Reilly"}
	published_years := []int64{2004, 2022, 2019, 2010, 2015}
	abstracts := []string{
		"An exploration of quantum mechanics.",
		"A magic journey through time and space.",
		"An introduction to machine learning techniques.",
		"A deep dive into the history of ancient civilizations.",
		"A fantasy tale of dragons and magic.",
	}

	for i := 0; i < len(titles); i++ {
		query := "CREATE (b:Book { id:$id,title: $title, title_embedding: $embeddings, published_year: $year, abstract:$abstract})"
		prepare, err := v.conn.Prepare(query)
		if err != nil {
			panic(err)
		}
		embeddings, err := v.vectorFunc(context.Background(), titles[i])
		if err != nil {
			panic(err)
		}
		_, err = v.conn.Execute(prepare, map[string]any{
			"id":         i + 1,
			"title":      titles[i],
			"embeddings": FloatListToAnyList(embeddings),
			"year":       published_years[i],
			"abstract":   abstracts[i],
		})
		if err != nil {
			panic(err)
		}

	}
	fmt.Println("insert book done")

	for i := 0; i < len(publishers); i++ {
		query := "CREATE (p:Publisher {name: $publisher})"
		prepare, err := v.conn.Prepare(query)
		if err != nil {
			panic(err)
		}
		_, err = v.conn.Execute(prepare, map[string]any{
			"publisher": publishers[i],
		})
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("insert publisher done")

	for i := 0; i < len(publishers); i++ {
		query := "MATCH (b:Book {title: $title})   MATCH (p:Publisher {name: $publisher})     CREATE (b)-[:PublishedBy]->(p);"
		prepare, err := v.conn.Prepare(query)
		if err != nil {
			panic(err)
		}
		_, err = v.conn.Execute(prepare, map[string]any{
			"title":     titles[i],
			"publisher": publishers[i],
		})
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("DONE")
	return v
}

func (v *VectorDB) CHECKPOINT() *VectorDB {
	_, err := v.conn.Query("CHECKPOINT")
	if err != nil {
		return nil
	}
	return v
}
