package snowflake

var node *Node

func SetNode(nodeId int64) {
	n, err := NewNode(nodeId)
	if err != nil {
		panic(err)
	}
	node = n
}

func Generate() ID {
	return node.Generate()
}
