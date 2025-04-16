package main

import (
	"github.com/philippgille/chromem-go"
)

const embeddingModelForZh = "smartcreation/bge-large-zh-v1.5"

func EmbeddingFunc() chromem.EmbeddingFunc {
	return chromem.NewEmbeddingFuncOllama(embeddingModelForZh, "")

}
