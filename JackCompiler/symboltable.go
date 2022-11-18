package main

type SymbolKind int

const (
	SYMBOL_NONE SymbolKind = iota
	SYMBOL_STATIC
	SYMBOL_FIELD
	SYMBOL_ARG
	SYMBOL_VAR
)

type Symbol struct {
	name       string
	symbolType string
	kind       SymbolKind
	index      int
}

type SymbolTable struct {
	table map[string]Symbol
	count map[SymbolKind]int
}

func NewSymbolTable() *SymbolTable {
	table := make(map[string]Symbol)

	count := map[SymbolKind]int{
		SYMBOL_STATIC: 0,
		SYMBOL_FIELD:  0,
		SYMBOL_ARG:    0,
		SYMBOL_VAR:    0,
	}

	s := SymbolTable{
		table: table,
		count: count,
	}

	return &s
}

func (s *SymbolTable) Reset() {
	s.table = make(map[string]Symbol)
	for kind := range s.count {
		s.count[kind] = 0
	}
}

func (s *SymbolTable) Define(name, symbolType string, kind SymbolKind) {
	if kind == SYMBOL_NONE {
		panic("SymbolTable.Define: invalid kind")
	}

	symbol := Symbol{
		name:       name,
		symbolType: symbolType,
		kind:       kind,
		index:      s.count[kind],
	}

	s.table[name] = symbol
	s.count[kind]++
}

func (s *SymbolTable) VarCount(kind SymbolKind) int {
	if count, ok := s.count[kind]; ok {
		return count
	}

	return 0
}

func (s *SymbolTable) KindOf(name string) SymbolKind {
	if symbol, ok := s.table[name]; ok {
		return symbol.kind
	}

	return SYMBOL_NONE
}

func (s *SymbolTable) TypeOf(name string) string {
	if symbol, ok := s.table[name]; ok {
		return symbol.symbolType
	}

	return ""
}

func (s *SymbolTable) IndexOf(name string) int {
	if symbol, ok := s.table[name]; ok {
		return symbol.index
	}

	return -1
}
