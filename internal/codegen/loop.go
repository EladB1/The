package codegen

import "github.com/EladB1/The/internal/irgen"

func handleBlock(block irgen.Block) Block {
	return Block{
		Label: block.Label,
		Code:  generateBody(block.Code),
	}
}

func handleLoop(block irgen.Loop) Loop {
	return Loop{
		Label: block.Label,
		Code:  generateBody(block.Code),
	}
}
