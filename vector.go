package kuzu_vector

import (
	"context"
	"fmt"
	"github.com/kuzudb/go-kuzu"
	"github.com/philippgille/chromem-go"
	"os"
)

type VectorDB struct {
	conn       *kuzu.Connection
	db         *kuzu.Database
	query      []string
	vectorFunc chromem.EmbeddingFunc
}

func (*VectorDB) Name() string {
	return "vector_db"
}

func (v *VectorDB) InitConn(needRemove bool) *VectorDB {
	dbPath := v.Name()
	if needRemove {
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
func (v *VectorDB) CreateVectorNode() *VectorDB {
	queries := []string{
		"CREATE NODE TABLE Book(id SERIAL PRIMARY KEY, title STRING, title_embedding FLOAT[1024], published_year INT64)",
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

func (v *VectorDB) CreateVectorIndex() *VectorDB {
	index := "CALL CREATE_VECTOR_INDEX('Book','title_vec_index','title_embedding')"

	queryRes, err := v.conn.Query(index)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println("Executing query:", index)
	fmt.Println("Query Result:", queryRes)

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

	for i := 0; i < len(titles); i++ {
		query := "CREATE (b:Book {title: $title, title_embedding: $embeddings, published_year: $year})"
		prepare, err := v.conn.Prepare(query)
		if err != nil {
			panic(err)
		}
		embeddings, err := v.vectorFunc(context.Background(), titles[i])
		if err != nil {
			panic(err)
		}
		execute, err := v.conn.Execute(prepare, map[string]any{
			"title":      titles[i],
			"embeddings": FloatListToAnyList(embeddings),
			"year":       published_years[i],
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(execute.ToString())

	}
	fmt.Println("insert book done")

	for i := 0; i < len(publishers); i++ {
		query := "CREATE (p:Publisher {name: $publisher})"
		prepare, err := v.conn.Prepare(query)
		if err != nil {
			panic(err)
		}
		execute, err := v.conn.Execute(prepare, map[string]any{
			"publisher": publishers[i],
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(execute.ToString())
	}

	fmt.Println("insert publisher done")

	for i := 0; i < len(publishers); i++ {
		query := "MATCH (b:Book {title: $title})   MATCH (p:Publisher {name: $publisher})     CREATE (b)-[:PublishedBy]->(p);"
		fmt.Println("Executing query:", query)
		prepare, err := v.conn.Prepare(query)
		if err != nil {
			panic(err)
		}
		execute, err := v.conn.Execute(prepare, map[string]any{
			"title":     titles[i],
			"publisher": publishers[i],
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(execute.ToString())
	}
	fmt.Println("DONE")
	return v
}

func GetKuzuVectorConn() (*kuzu.Database, *kuzu.Connection) {
	dbPath := "vector_db"
	//os.RemoveAll(dbPath)
	// Open a database with default system configuration.
	systemConfig := kuzu.DefaultSystemConfig()
	systemConfig.BufferPoolSize = 1024 * 1024 * 1024
	db, err := kuzu.OpenDatabase(dbPath, systemConfig)
	if err != nil {
		panic(err)
	}
	//defer db.Close()

	// Open a connection to the database.
	conn, err := kuzu.OpenConnection(db)
	if err != nil {
		panic(err)
	}

	return db, conn
}

func LoadVector() {
	_, conn := GetKuzuVectorConn()
	installPre, err := conn.Prepare("INSTALL VECTOR;")
	if err != nil {
		return
	}
	executeInstall, err := conn.Execute(installPre, nil)
	if err != nil {
		return
	}
	fmt.Println(executeInstall.ToString())

	loadPre, err := conn.Prepare("LOAD VECTOR;")
	if err != nil {
		return
	}
	executeLoad, err := conn.Execute(loadPre, nil)
	if err != nil {
		return
	}
	fmt.Println(executeLoad.ToString())
}

func QueryVector() {
	_, conn := GetKuzuVectorConn()

	f := EmbeddingFunc()
	vector, err := f(context.Background(), "quantum machine learning")
	if err != nil {
		panic(err)
	}
	pre, err := conn.Prepare("  CALL QUERY_VECTOR_INDEX('Book','title_vec_index',$query_vector,2 )RETURN node.title ORDER BY distance;")
	if err != nil {
		panic(err)
	}
	execute, err := conn.Execute(pre, map[string]any{
		"query_vector": vector,
	})
	if err != nil {
		return
	}
	fmt.Println(execute.ToString())

}

func FloatListToAnyList(floats []float32) []any {
	var list []any
	for _, v := range floats {
		list = append(list, v)
	}
	return list
}
