package main

import (
	"fmt"
	"testing"
)

func TestReadForFTS(t *testing.T) {
	read := &Read{}
	read.OnStart()
	defer read.OnEnd()

	read.QueryEntityForFTS("MCP")
	fmt.Println("1111")
	read.QueryRelationShipForFTS("mcp")

}

func TestReadEntity(t *testing.T) {
	read := &Read{}
	read.OnStart()
	defer read.OnEnd()

	read.QueryEntityByName([]string{"MCP", "Eino"}...)
}

func TestReadForVector(t *testing.T) {
	read := &Read{}
	read.OnStart()
	defer read.OnEnd()

	read.QueryEntityForVector("MCP是什么东西，能做什么")
	fmt.Println("1111")
	read.QueryRelationShipForVector("MCP是什么东西，能做什么")
}
