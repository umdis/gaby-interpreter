package parser

import (
	"bytes"
	"strings"

	"github.com/usuario/gaby-interpreter/internal/lexer"
)

// Node es la interfaz base para todos los nodos del AST
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement es la interfaz para todos los nodos de tipo sentencia
type Statement interface {
	Node
	statementNode()
}

// Expression es la interfaz para todos los nodos de tipo expresión
type Expression interface {
	Node
	expressionNode()
}

// Program es el nodo raíz del AST
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// Identificador representa un identificador (variable, función, etc.)
type Identifier struct {
	Token lexer.Token // token IDENT
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// ExpressionStatement es una sentencia que contiene una expresión
type ExpressionStatement struct {
	Token      lexer.Token // la primera token de la expresión
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// LetStatement representa una sentencia de asignación/declaración de variable (guarda)
type LetStatement struct {
	Token lexer.Token // token VAR
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

// ReturnStatement representa una sentencia de retorno (devolver)
type ReturnStatement struct {
	Token       lexer.Token // token RETURN
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

// BlockStatement representa un bloque de código (entre llaves)
type BlockStatement struct {
	Token      lexer.Token // token {
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	out.WriteString("{ ")
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	out.WriteString(" }")

	return out.String()
}

// IntegerLiteral representa un literal numérico entero
type IntegerLiteral struct {
	Token lexer.Token // token NUM
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral representa un literal numérico decimal
type FloatLiteral struct {
	Token lexer.Token // token NUM
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral representa un literal de cadena
type StringLiteral struct {
	Token lexer.Token // token STRING
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

// BooleanLiteral representa un literal booleano (verdad/falso)
type BooleanLiteral struct {
	Token lexer.Token // token TRUE o FALSE
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// NullLiteral representa un literal nulo
type NullLiteral struct {
	Token lexer.Token // token NULL
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NullLiteral) String() string       { return "nulo" }

// PrefixExpression representa una expresión prefija (ej. -5, !true)
type PrefixExpression struct {
	Token    lexer.Token // El token de operación prefija (!, -, etc.)
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

// InfixExpression representa una expresión infija (ej. 5 + 5, 10 == 10)
type InfixExpression struct {
	Token    lexer.Token // El token de operación ('+', '==', etc.)
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

// IfExpression representa una expresión condicional (si/sino)
type IfExpression struct {
	Token       lexer.Token // token IF
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("si ")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		out.WriteString(" sino ")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}

// WhileExpression representa un bucle mientras
type WhileExpression struct {
	Token     lexer.Token // token WHILE
	Condition Expression
	Body      *BlockStatement
}

func (we *WhileExpression) expressionNode()      {}
func (we *WhileExpression) TokenLiteral() string { return we.Token.Literal }
func (we *WhileExpression) String() string {
	var out bytes.Buffer

	out.WriteString("mientras ")
	out.WriteString(we.Condition.String())
	out.WriteString(" ")
	out.WriteString(we.Body.String())

	return out.String()
}

// ForExpression representa un bucle para
type ForExpression struct {
	Token     lexer.Token // token FOR
	Init      Statement
	Condition Expression
	Update    Statement
	Body      *BlockStatement
}

func (fe *ForExpression) expressionNode()      {}
func (fe *ForExpression) TokenLiteral() string { return fe.Token.Literal }
func (fe *ForExpression) String() string {
	var out bytes.Buffer

	out.WriteString("para ")
	if fe.Init != nil {
		out.WriteString(fe.Init.String())
	}
	out.WriteString("; ")
	if fe.Condition != nil {
		out.WriteString(fe.Condition.String())
	}
	out.WriteString("; ")
	if fe.Update != nil {
		out.WriteString(fe.Update.String())
	}
	out.WriteString(" ")
	out.WriteString(fe.Body.String())

	return out.String()
}

// FunctionLiteral representa una definición de función (fun)
type FunctionLiteral struct {
	Token      lexer.Token // token FUNCTION
	Parameters []*Identifier
	Body       *BlockStatement
	Name       string
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	if fl.Name != "" {
		out.WriteString(" " + fl.Name)
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}

// CallExpression representa una llamada a función
type CallExpression struct {
	Token     lexer.Token // token (
	Function  Expression  // Puede ser identificador o expresión de función
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

// IndexExpression representa acceso a elementos de arrays (ej. array[0])
type IndexExpression struct {
	Token lexer.Token // El token '['
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")

	return out.String()
}

// ArrayLiteral representa un literal de array (lista)
type ArrayLiteral struct {
	Token    lexer.Token // El token '['
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

// HashLiteral representa un literal de mapa (diccionario)
type HashLiteral struct {
	Token lexer.Token // El token '{'
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+": "+value.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

// DotExpression representa una expresión de acceso a atributo mediante punto (objeto.atributo)
type DotExpression struct {
	Token    lexer.Token // El token '.'
	Object   Expression
	Property *Identifier
}

func (de *DotExpression) expressionNode()      {}
func (de *DotExpression) TokenLiteral() string { return de.Token.Literal }
func (de *DotExpression) String() string {
	var out bytes.Buffer

	out.WriteString(de.Object.String())
	out.WriteString(".")
	out.WriteString(de.Property.String())

	return out.String()
}

// ClassLiteral representa una declaración de clase
type ClassLiteral struct {
	Token       lexer.Token // token CLASS
	Name        *Identifier
	Parent      *Identifier
	Interfaces  []*Identifier
	Properties  []*LetStatement
	Methods     []*FunctionLiteral
}

func (cl *ClassLiteral) expressionNode()      {}
func (cl *ClassLiteral) TokenLiteral() string { return cl.Token.Literal }
func (cl *ClassLiteral) String() string {
	var out bytes.Buffer

	out.WriteString("clase ")
	out.WriteString(cl.Name.String())
	
	if cl.Parent != nil {
		out.WriteString(" hereda ")
		out.WriteString(cl.Parent.String())
	}
	
	if len(cl.Interfaces) > 0 {
		out.WriteString(" implementa ")
		ints := []string{}
		for _, intf := range cl.Interfaces {
			ints = append(ints, intf.String())
		}
		out.WriteString(strings.Join(ints, ", "))
	}
	
	out.WriteString(" {\n")
	
	for _, prop := range cl.Properties {
		out.WriteString("  " + prop.String() + "\n")
	}
	
	for _, method := range cl.Methods {
		out.WriteString("  " + method.String() + "\n")
	}
	
	out.WriteString("}")

	return out.String()
}

// NewExpression representa una creación de objeto mediante 'nuevo'
type NewExpression struct {
	Token     lexer.Token // token NEW
	Class     Expression
	Arguments []Expression
}

func (ne *NewExpression) expressionNode()      {}
func (ne *NewExpression) TokenLiteral() string { return ne.Token.Literal }
func (ne *NewExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ne.Arguments {
		args = append(args, a.String())
	}

	out.WriteString("nuevo ")
	out.WriteString(ne.Class.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}