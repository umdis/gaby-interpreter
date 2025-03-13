package lexer

// TokenType es el tipo de un token
type TokenType string

// Token representa un token en el código fuente
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// Constantes para los tipos de tokens
const (
	ILLEGAL = "ILLEGAL" // Token no reconocido
	EOF     = "EOF"     // Fin de archivo

	// Identificadores y literales
	IDENT  = "IDENT"  // identificadores: x, y, foo, etc.
	NUM    = "NUM"    // números: 1343456, 1.34, etc.
	STRING = "STRING" // cadenas: "foo", "bar", etc.

	// Operadores
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	MOD      = "%"
	POWER    = "^"

	LT     = "<"
	GT     = ">"
	EQ     = "=="
	NOT_EQ = "!="

	// Asignaciones compuestas
	PLUS_ASSIGN  = "+="
	MINUS_ASSIGN = "-="
	MUL_ASSIGN   = "*="
	DIV_ASSIGN   = "/="
	MOD_ASSIGN   = "%="
	POW_ASSIGN   = "^="
	DECLARE      = ":="

	// Delimitadores
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	DOT       = "."

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Palabras clave
	FUNCTION  = "FUNCTION"
	CLASS     = "CLASS"
	PROTO     = "PROTO"
	IF        = "IF"
	ELSE      = "ELSE"
	WHEN      = "WHEN"
	RETURN    = "RETURN"
	TRUE      = "TRUE"
	FALSE     = "FALSE"
	NULL      = "NULL"
	WHILE     = "WHILE"
	FOR       = "FOR"
	REPEAT    = "REPEAT"
	DO        = "DO"
	BREAK     = "BREAK"
	CONTINUE  = "CONTINUE"
	SWITCH    = "SWITCH"
	CASE      = "CASE"
	DEFAULT   = "DEFAULT"
	IN        = "IN"
	FROM      = "FROM"
	TO        = "TO"
	AND       = "AND"
	OR        = "OR"
	NOT       = "NOT"
	IS        = "IS"
	ISNOT     = "ISNOT"
	NEW       = "NEW"
	EXTENDS   = "EXTENDS"
	IMPLEMENTS = "IMPLEMENTS"
	VAR       = "VAR"
	PUBLIC    = "PUBLIC"
	PRIVATE   = "PRIVATE"
	PROTECTED = "PROTECTED"
	STATIC    = "STATIC"
	FINAL     = "FINAL"
	THIS      = "THIS"
	SUPER     = "SUPER"
	TRY       = "TRY"
	CATCH     = "CATCH"
	FINALLY   = "FINALLY"
	THROW     = "THROW"
)

// Mapeo de palabras clave a tipos de tokens
var keywords = map[string]TokenType{
	"fun":        FUNCTION,
	"clase":      CLASS,
	"proto":      PROTO,
	"si":         IF,
	"sino":       ELSE,
	"cuando":     WHEN,
	"devolver":   RETURN,
	"verdad":     TRUE,
	"falso":      FALSE,
	"nulo":       NULL,
	"mientras":   WHILE,
	"para":       FOR,
	"repetir":    REPEAT,
	"haz":        DO,
	"romper":     BREAK,
	"continuar":  CONTINUE,
	"evaluar":    SWITCH,
	"caso":       CASE,
	"defecto":    DEFAULT,
	"en":         IN,
	"desde":      FROM,
	"hasta":      TO,
	"y":          AND,
	"o":          OR,
	"no":         NOT,
	"es":         IS,
	"no_es":      ISNOT,
	"nuevo":      NEW,
	"extiende":   EXTENDS,
	"implementa": IMPLEMENTS,
	"guarda":     VAR,
	"publico":    PUBLIC,
	"privado":    PRIVATE,
	"protegido":  PROTECTED,
	"estatico":   STATIC,
	"final":      FINAL,
	"esto":       THIS,
	"super":      SUPER,
	"intentar":   TRY,
	"atrapar":    CATCH,
	"finalmente": FINALLY,
	"lanzar":     THROW,
}

// LookupIdent revisa si un identificador es una palabra clave.
// Si lo es, devuelve el tipo de token de la palabra clave.
// Si no, devuelve IDENT.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}