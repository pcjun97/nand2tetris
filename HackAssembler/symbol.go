package main

type Symbol struct {
	table map[string]int
}

func NewSymbol() *Symbol {
	table := map[string]int{
		"R0":     0,
		"R1":     1,
		"R2":     2,
		"R3":     3,
		"R4":     4,
		"R5":     5,
		"R6":     6,
		"R7":     7,
		"R8":     8,
		"R9":     9,
		"R10":    10,
		"R11":    11,
		"R12":    12,
		"R13":    13,
		"R14":    14,
		"R15":    15,
		"SP":     0,
		"LCL":    1,
		"ARG":    2,
		"THIS":   3,
		"THAT":   4,
		"SCREEN": 16384,
		"KBD":    24576,
	}

	s := Symbol{table: table}
	return &s
}

func (s *Symbol) AddEntry(symbol string, address int) {
	s.table[symbol] = address
}

func (s *Symbol) Contains(symbol string) bool {
	_, ok := s.table[symbol]
	return ok
}

func (s *Symbol) GetAddress(symbol string) (int, bool) {
	address, ok := s.table[symbol]
	return address, ok
}
