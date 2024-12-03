package types

type Sheet struct {
	APIKey        string      `yaml:"api_key"`
	Men           SheetEntry  `yaml:"men"`
	Women         SheetEntry  `yaml:"women"`
	CellTypeColor SheetColors `yaml:"cell_type_color"`
}

type SheetColors struct {
	Home RGB `yaml:"home"`
	Away RGB `yaml:"away"`
	Cup  RGB `yaml:"cup"`
}

type SheetEntry struct {
	ID string `yaml:"id"`
}

type SheetCellData struct {
	Value      string
	Background RGB
}
