package stdlib

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/usuario/gaby-interpreter/internal/object"
)

// LoadStdlib carga las funciones de la biblioteca estándar en el entorno
func LoadStdlib(env *object.Environment) {
	// Funciones de E/S
	registerBuiltin(env, "mostrar", mostrar)
	registerBuiltin(env, "leer", leer)
	registerBuiltin(env, "leer_numero", leerNumero)
	
	// Funciones matemáticas
	registerBuiltin(env, "abs", abs)
	registerBuiltin(env, "redondear", redondear)
	registerBuiltin(env, "piso", piso)
	registerBuiltin(env, "techo", techo)
	registerBuiltin(env, "potencia", potencia)
	registerBuiltin(env, "raiz", raiz)
	
	// Funciones de texto
	registerBuiltin(env, "texto", convertirATexto)
	registerBuiltin(env, "num", convertirANumero)
	registerBuiltin(env, "mayusculas", mayusculas)
	registerBuiltin(env, "minusculas", minusculas)
	registerBuiltin(env, "recortar", recortar)
	registerBuiltin(env, "contiene", contiene)
	registerBuiltin(env, "reemplazar", reemplazar)
	registerBuiltin(env, "dividir", dividir)
	
	// Funciones de tiempo
	registerBuiltin(env, "ahora", ahora)
	registerBuiltin(env, "dormir", dormir)
	
	// Funciones de sistema
	registerBuiltin(env, "args", args)
	registerBuiltin(env, "salir", salir)
	registerBuiltin(env, "cargar", cargar)
	
	// Funciones de colecciones
	registerBuiltin(env, "longitud", longitud)
	registerBuiltin(env, "agregar", agregar)
	registerBuiltin(env, "eliminar", eliminar)
	registerBuiltin(env, "rango", rango)
}

// registerBuiltin registra una función incorporada en el entorno
func registerBuiltin(env *object.Environment, name string, fn object.BuiltinFunction) {
	env.Set(name, &object.Builtin{Fn: fn})
}

// Funciones de E/S

func mostrar(args ...object.Object) object.Object {
	for _, arg := range args {
		fmt.Println(arg.Inspect())
	}
	return &object.Null{}
}

func leer(args ...object.Object) object.Object {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return newError("error al leer entrada: %s", err)
	}
	
	// Eliminar salto de línea final
	input = strings.TrimRight(input, "\r\n")
	
	return &object.String{Value: input}
}

func leerNumero(args ...object.Object) object.Object {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return newError("error al leer entrada: %s", err)
	}
	
	// Eliminar salto de línea final y espacios
	input = strings.TrimSpace(input)
	
	// Intentar convertir a entero
	if intVal, err := strconv.ParseInt(input, 10, 64); err == nil {
		return &object.Integer{Value: intVal}
	}
	
	// Intentar convertir a flotante
	if floatVal, err := strconv.ParseFloat(input, 64); err == nil {
		return &object.Float{Value: floatVal}
	}
	
	return newError("no se pudo convertir '%s' a número", input)
}

// Funciones matemáticas

func abs(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("número incorrecto de argumentos: se esperaba 1, se obtuvo %d", len(args))
	}
	
	switch arg := args[0].(type) {
	case *object.Integer:
		value := arg.Value
		if value < 0 {
			value = -value
		}
		return &object.Integer{Value: value}
	case *object.Float:
		return &object.Float{Value: math.Abs(arg.Value)}
	default:
		return newError("argumento no válido para 'abs': %s", args[0].Type())
	}
}

func redondear(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("número incorrecto de argumentos: se esperaba 1, se obtuvo %d", len(args))
	}
	
	switch arg := args[0].(type) {
	case *object.Integer:
		return arg // Un entero ya está redondeado
	case *object.Float:
		return &object.Float{Value: math.Round(arg.Value)}
	default:
		return newError("argumento no válido para 'redondear': %s", args[0].Type())
	}
}

func piso(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("número incorrecto de argumentos: se esperaba 1, se obtuvo %d", len(args))
	}
	
	switch arg := args[0].(type) {
	case *object.Integer:
		return arg // Un entero ya está redondeado hacia abajo
	case *object.Float:
		return &object.Float{Value: math.Floor(arg.Value)}
	default:
		return newError("argumento no válido para 'piso': %s", args[0].Type())
	}
}

func techo(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("número incorrecto de argumentos: se esperaba 1, se obtuvo %d", len(args))
	}
	
	switch arg := args[0].(type) {
	case *object.Integer:
		return arg // Un entero ya está redondeado hacia arriba
	case *object.Float:
		return &object.Float{Value: math.Ceil(arg.Value)}
	default:
		return newError("argumento no válido para 'techo': %s", args[0].Type())
	}
}

func potencia(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("número incorrecto de argumentos: se esperaba 2, se obtuvo %d", len(args))
	}
	
	var base, exp float64
	
	switch arg := args[0].(type) {
	case *object.Integer:
		base = float64(arg.Value)
	case *object.Float:
		base = arg.Value
	default:
		return newError("primer argumento no válido para 'potencia': %s", args[0].Type())
	}
	
	switch arg := args[1].(type) {
	case *object.Integer:
		exp = float64(arg.Value)
	case *object.Float:
		exp = arg.Value
	default:
		return newError("segundo argumento no válido para 'potencia': %s", args[1].Type())
	}
	
	result := math.Pow(base, exp)
	
	// Si el resultado es un entero exacto, devolver entero
	if result == math.Floor(result) && result <= float64(math.MaxInt64) && result >= float64(math.MinInt64) {
		return &object.Integer{Value: int64(result)}
	}
	
	return &object.Float{Value: result}
}

func raiz(args ...object.Object) object.Object {
	if len(args) != 1 && len(args) != 2 {
		return newError("número incorrecto de argumentos: se esperaba 1 o 2, se obtuvo %d", len(args))
	}
	
	var value float64
	
	switch arg := args[0].(type) {
	case *object.Integer:
		value = float64(arg.Value)
	case *object.Float:
		value = arg.Value
	default:
		return newError("primer argumento no válido para 'raiz': %s", args[0].Type())
	}
	
	if value < 0 {
		return newError("no se puede calcular la raíz de un número negativo")
	}
	
	// Si se proporciona el segundo argumento, es el índice de la raíz
	if len(args) == 2 {
		var indice float64
		
		switch arg := args[1].(type) {
		case *object.Integer:
			indice = float64(arg.Value)
		case *object.Float:
			indice = arg.Value
		default:
			return newError("segundo argumento no válido para 'raiz': %s", args[1].Type())
		}
		
		if indice == 0 {
			return newError("el índice de la raíz no puede ser cero")
		}
		
		result := math.Pow(value, 1/indice)
		return &object.Float{Value: result}
	}
	
	// Por defecto, raíz cuadrada
	return &object.Float{Value: math.Sqrt(value)}
}

// Funciones de texto

func convertirATexto(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("número incorrecto de argumentos: se esperaba 1, se obtuvo %d", len(args))
	}
	
	return &object.String{Value: args[0].Inspect()}
}

func convertirANumero(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("número incorrecto de argumentos: se esperaba 1, se obtuvo %d", len(args))
	}
	
	switch arg := args[0].(type) {
	case *object.Integer:
		return arg
	case *object.Float:
		return arg
	case *object.String:
		// Intentar convertir a entero
		if intVal, err := strconv.ParseInt(arg.Value, 10, 64); err == nil {
			return &object.Integer{Value: intVal}
		}
		
		// Intentar convertir a flotante
		if floatVal, err := strconv.ParseFloat(arg.Value, 64); err == nil {
			return &object.Float{Value: floatVal}
		}
		
		return newError("no se pudo convertir '%s' a número", arg.Value)
	default:
		return newError("argumento no válido para 'num': %s", args[0].Type())
	}
}

func mayusculas(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("número incorrecto de argumentos: se esperaba 1, se obtuvo %d", len(args))
	}
	
	if arg, ok := args[0].(*object.String); ok {
		return &object.String{Value: strings.ToUpper(arg.Value)}
	}
	
	return newError("argumento no válido para 'mayusculas': %s", args[0].Type())
}

func minusculas(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("número incorrecto de argumentos: se esperaba 1, se obtuvo %d", len(args))
	}
	
	if arg, ok := args[0].(*object.String); ok {
		return &object.String{Value: strings.ToLower(arg.Value)}
	}
	
	return newError("argumento no válido para 'minusculas': %s", args[0].Type())
}

func recortar(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("número incorrecto de argumentos: se esperaba 1, se obtuvo %d", len(args))
	}
	
	if arg, ok := args[0].(*object.String); ok {
		return &object.String{Value: strings.TrimSpace(arg.Value)}
	}
	
	return newError("argumento no válido para 'recortar': %s", args[0].Type())
}

func contiene(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("número incorrecto de argumentos: se esperaba 2, se obtuvo %d", len(args))
	}
	
	if str, ok := args[0].(*object.String); ok {
		if substr, ok := args[1].(*object.String); ok {
			if strings.Contains(str.Value, substr.Value) {
				return TRUE
			}
			return FALSE
		}
	}
	
	return newError("argumentos no válidos para 'contiene': %s, %s", args[0].Type(), args[1].Type())
}

func reemplazar(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("número incorrecto de argumentos: se esperaba 3, se obtuvo %d", len(args))
	}
	
	if str, ok := args[0].(*object.String); ok {
		if old, ok := args[1].(*object.String); ok {
			if new, ok := args[2].(*object.String); ok {
				return &object.String{Value: strings.ReplaceAll(str.Value, old.Value, new.Value)}
			}
		}
	}
	
	return newError("argumentos no válidos para 'reemplazar'")
}

func dividir(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("número incorrecto de argumentos: se esperaba 2, se obtuvo %d", len(args))
	}
	
	if str, ok := args[0].(*object.String); ok {
		if sep, ok := args[1].(*object.String); ok {
			parts := strings.Split(str.Value, sep.Value)
			elements := make([]object.Object, len(parts))
			for i, part := range parts {
				elements[i] = &object.String{Value: part}
			}
			return &object.Array{Elements: elements}
		}
	}
	
	return newError("argumentos no válidos para 'dividir': %s, %s", args[0].Type(), args[1].Type())
}

// Funciones de tiempo

func ahora(args ...object.Object) object.Object {
	if len(args) != 0 {
		return newError("número incorrecto de argumentos: se esperaba 0, se obtuvo %d", len(args))
	}
	
	now := time.Now()
	return &object.String{Value: now.Format("2006-01-02 15:04:05")}
}

func dormir(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("número incorrecto de argumentos: se esperaba 1, se obtuvo %d", len(args))
	}
	
	var duracion float64
	
	switch arg := args[0].(type) {
	case *object.Integer:
		duracion = float64(arg.Value)
	case *object.Float:
		duracion = arg.Value
	default:
		return newError("argumento no válido para 'dormir': %s", args[0].Type())
	}
	
	time.Sleep(time.Duration(duracion * float64(time.Second)))
	return &object.Null{}
}

// Funciones de sistema

func args(args ...object.Object) object.Object {
	if len(args) != 0 {
		return newError("número incorrecto de argumentos: se esperaba 0, se obtuvo %d", len(args))
	}
	
	osArgs := os.Args[1:]
	elements := make([]object.Object, len(osArgs))
	
	for i, arg := range osArgs {
		elements[i] = &object.String{Value: arg}
	}
	
	return &object.Array{Elements: elements}
}

func salir(args ...object.Object) object.Object {
	code := 0
	
	if len(args) == 1 {
		switch arg := args[0].(type) {
		case *object.Integer:
			code = int(arg.Value)
		default:
			return newError("argumento no válido para 'salir': %s", args[0].Type())
		}
	} else if len(args) > 1 {
		return newError("número incorrecto de argumentos: se esperaba 0 o 1, se obtuvo %d", len(args))
	}
	
	os.Exit(code)
	return &object.Null{} // Nunca se llega aquí
}

func cargar(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("número incorrecto de argumentos: se esperaba 1, se obtuvo %d", len(args))
	}
	
	if filepath, ok := args[0].(*object.String); ok {
		// Verificar extensión
		if !strings.HasSuffix(filepath.Value, ".gaby") {
			return newError("el archivo debe tener extensión .gaby")
		}
		
		// Leer contenido del archivo
		content, err := os.ReadFile(filepath.Value)
		if err != nil {
			return newError("error al leer el archivo: %s", err)
		}
		
		// Retornar el contenido como string (para que el programa principal lo evalúe)
		return &object.String{Value: string(content)}
	}
	
	return newError("argumento no válido para 'cargar': %s", args[0].Type())
}

// Funciones de colecciones

func longitud(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("número incorrecto de argumentos: se esperaba 1, se obtuvo %d", len(args))
	}
	
	switch arg := args[0].(type) {
	case *object.String:
		return &object.Integer{Value: int64(len(arg.Value))}
	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}
	case *object.Hash:
		return &object.Integer{Value: int64(len(arg.Pairs))}
	default:
		return newError("argumento no válido para 'longitud': %s", args[0].Type())
	}
}

func agregar(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("número incorrecto de argumentos: se esperaba 2, se obtuvo %d", len(args))
	}
	
	if arr, ok := args[0].(*object.Array); ok {
		newElements := make([]object.Object, len(arr.Elements))
		copy(newElements, arr.Elements)
		newElements = append(newElements, args[1])
		return &object.Array{Elements: newElements}
	}
	
	return newError("primer argumento no válido para 'agregar': %s", args[0].Type())
}

func eliminar(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("número incorrecto de argumentos: se esperaba 2, se obtuvo %d", len(args))
	}
	
	if arr, ok := args[0].(*object.Array); ok {
		if idx, ok := args[1].(*object.Integer); ok {
			i := idx.Value
			if i < 0 || i >= int64(len(arr.Elements)) {
				return newError("índice fuera de rango")
			}
			
			newElements := make([]object.Object, 0, len(arr.Elements)-1)
			newElements = append(newElements, arr.Elements[:i]...)
			newElements = append(newElements, arr.Elements[i+1:]...)
			
			return &object.Array{Elements: newElements}
		}
	}
	
	return newError("argumentos no válidos para 'eliminar': %s, %s", args[0].Type(), args[1].Type())
}

func rango(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("número incorrecto de argumentos: se esperaba 2, se obtuvo %d", len(args))
	}
	
	var inicio, fin int64
	
	switch arg := args[0].(type) {
	case *object.Integer:
		inicio = arg.Value
	default:
		return newError("primer argumento no válido para 'rango': %s", args[0].Type())
	}
	
	switch arg := args[1].(type) {
	case *object.Integer:
		fin = arg.Value
	default:
		return newError("segundo argumento no válido para 'rango': %s", args[1].Type())
	}
	
	if inicio > fin {
		return newError("el inicio no puede ser mayor que el fin")
	}
	
	elements := make([]object.Object, 0, fin-inicio+1)
	for i := inicio; i <= fin; i++ {
		elements = append(elements, &object.Integer{Value: i})
	}
	
	return &object.Array{Elements: elements}
}

// Constantes y utilidades

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}