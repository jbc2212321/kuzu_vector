package main

func main() {
	c := &Create{}

	c.OnStart(true)

	c.CreateNode().InsertNode().
		CreateFTSIndex("Entity").CreateFTSIndex("EntityRelationship").                                                         // 给两个节点都创建全文索引
		CreateVectorIndex("Entity", "description_embedding").CreateVectorIndex("EntityRelationship", "description_embedding"). // 给两个节点都创建向量索引
		CHECKPOINT()                                                                                                           // 目前kuzu不会通过wsl落库，需要手动执行下CHECKPOINT

	c.OnEnd()

	// 查询见 read_test
}
