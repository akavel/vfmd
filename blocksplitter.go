package vfmd

type BlockSplitter struct {
}

func (b *BlockSplitter) WriteLine(line []byte) error {
	// TODO(akavel): NIY
	return nil
}

func (b *BlockSplitter) Close() error {
	// TODO(akavel): NIY
	return nil
}

type Block interface {
	DetectBlock(startLine, secondLine []byte) bool
	Add(line []byte) (refluxLines [][]byte)
	Lines() [][]byte
}

type NullBlock struct{}
type ReferenceResolutionBlock struct{}
type SetextHeaderBlock struct{}
type CodeBlock struct{}
type AtxHeaderBlock struct{}
type QuoteBlock struct{}
type HorizontalRuleBlock struct{}
type UnorderedListBlock struct{}
type OrderedListBlock struct{}
type ParagraphBlock struct{}
