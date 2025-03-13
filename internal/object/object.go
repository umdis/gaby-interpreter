package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/usuario/gaby-interpreter/internal/parser"
)

// ObjectType es el tipo de un objeto
type ObjectType string

// Constantes para los tipos de objetos
const (
	INTEGER_OBJ      = "ENTERO"
	FLOAT_OBJ        = "DECIMAL"
	BOOLEAN_OBJ      = "BOOLEANO"
	NULL_OBJ         = "NULO"
	RETURN_VALUE_OBJ = "RETORNO"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCION"
	STRING_OBJ       = "TEXTO"
	BUILTIN_OBJ      = "INCORPORADO"
	ARRAY_OBJ        = "LISTA"
	HASH_OBJ         = "MAPA"
	CLASS_OBJ        = "CLASE"
	INSTANCE_OBJ     = "INSTANCIA"
)

// Object es la interfaz básica para todos los objetos
type Object interface {
	Type() ObjectType
	Inspect() string
}

// Integer representa un objeto entero
type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

// Float representa un objeto decimal
type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string  { return fmt.Sprintf("%g", f.Value) }

// Boolean representa un objeto booleano
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string {
	if b.Value {
		return "verdad"
	}
	return "falso"
}

// Null representa un objeto nulo
type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "nulo" }

// ReturnValue representa un valor de retorno de una función
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

// Error representa un objeto de error
type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

// Function representa un objeto función
type Function struct {
	Parameters []*parser.Identifier
	Body       *parser.BlockStatement
	Env        *Environment
	Name       string
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fun")
	if f.Name != "" {
		out.WriteString(" " + f.Name)
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

// String representa un objeto cadena de texto
type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

// BuiltinFunction es el tipo para funciones incorporadas
type BuiltinFunction func(args ...Object) Object

// Builtin representa un objeto de función incorporada
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "función incorporada" }

// Array representa un objeto lista/array
type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range a.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

// HashKey es el tipo para las claves de mapas
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// Hashable es la interfaz para objetos que pueden ser usados como claves en un mapa
type Hashable interface {
	HashKey() HashKey
}

// Implementaciones de Hashable para tipos básicos
func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	} else {
		value = 0
	}
	return HashKey{Type: b.Type(), Value: value}
}

func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

// HashPair representa un par clave-valor en un mapa
type HashPair struct {
	Key   Object
	Value Object
}

// Hash representa un objeto mapa/diccionario
type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

// Class representa un objeto clase
type Class struct {
	Name       string
	Properties map[string]Object
	Methods    map[string]*Function
	Parent     *Class
}

func (c *Class) Type() ObjectType { return CLASS_OBJ }
func (c *Class) Inspect() string {
	var out bytes.Buffer

	out.WriteString("clase ")
	out.WriteString(c.Name)

	if c.Parent != nil {
		out.WriteString(" hereda ")
		out.WriteString(c.Parent.Name)
	}

	out.WriteString(" { ... }")

	return out.String()
}

// Instance representa una instancia de una clase
type Instance struct {
	Class      *Class
	Properties map[string]Object
	Env        *Environment
}

func (i *Instance) Type() ObjectType { return INSTANCE_OBJ }
func (i *Instance) Inspect() string {
	return fmt.Sprintf("instancia de %s", i.Class.Name)
}

// Environment representa el entorno de ejecución
type Environment struct {
	store map[string]Object
	outer *Environment
}

// NewEnvironment crea un nuevo entorno
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

// NewEnclosedEnvironment crea un nuevo entorno con un entorno externo
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get obtiene un objeto del entorno
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set establece un objeto en el entorno
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}