package entity

// Position represents a position in source code
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Offset int `json:"offset"`
}

// CodeNode represents a parsed code element (function, struct, interface, etc.)
type CodeNode struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"` // function, struct, interface, type, variable, constant
	File     string                 `json:"file"`
	Package  string                 `json:"package"`
	Body     string                 `json:"body"`
	Position *Position              `json:"position,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Parameter represents a function parameter
type Parameter struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Return represents a function return value
type Return struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// FunctionCall represents a function call within code
type FunctionCall struct {
	Name      string    `json:"name"`
	Arguments []string  `json:"arguments"`
	Position  *Position `json:"position,omitempty"`
}

// FunctionInfo contains detailed information about a function
type FunctionInfo struct {
	Name       string          `json:"name"`
	Receiver   string          `json:"receiver,omitempty"`
	IsMethod   bool            `json:"is_method"`
	Parameters []*Parameter    `json:"parameters"`
	Returns    []*Return       `json:"returns"`
	Body       string          `json:"body"`
	Position   *Position       `json:"position,omitempty"`
	CallsTo    []*FunctionCall `json:"calls_to"`
	UsedTypes  []string        `json:"used_types"`
}

// StructField represents a field in a struct
type StructField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Tag      string `json:"tag,omitempty"`
	Embedded bool   `json:"embedded"`
}

// StructInfo contains information about a struct
type StructInfo struct {
	Name   string         `json:"name"`
	Fields []*StructField `json:"fields"`
	Body   string         `json:"body"`
}

// InterfaceMethod represents a method in an interface
type InterfaceMethod struct {
	Name       string       `json:"name"`
	Parameters []*Parameter `json:"parameters"`
	Returns    []*Return    `json:"returns"`
}

// InterfaceInfo contains information about an interface
type InterfaceInfo struct {
	Name    string             `json:"name"`
	Methods []*InterfaceMethod `json:"methods"`
	Body    string             `json:"body"`
}

// TypeInfo contains information about a type definition
type TypeInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Body string `json:"body"`
}

// VariableInfo contains information about a variable
type VariableInfo struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value,omitempty"`
	Body  string `json:"body"`
}

// ConstantInfo contains information about a constant
type ConstantInfo struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	Body  string `json:"body"`
}
