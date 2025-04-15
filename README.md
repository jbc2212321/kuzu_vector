



```
go get	"github.com/kuzudb/go-kuzu"
go get	"github.com/philippgille/chromem-go"
```


1. The TestVector method initializes kuzudb, including creating nodes, data, and vector indexes

2. The TestVectorSearch method implements vector search, but during execution, it is found that the index does not exist continuously

3. The TestMatch method attempted to view the index, but the execution result was none

4. The TestAddVectorIndex method attempts to add a vector index separately, but when it is executed, it tells me that the index exists, which makes me very confused

5. The EmbeddingFunc method converts text into a [] float32 array, with the underlying layer using ollama, which can be easily replaced in the testing environment

