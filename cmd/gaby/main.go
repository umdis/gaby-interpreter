package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/usuario/gaby-interpreter/internal/evaluator"
	"github.com/usuario/gaby-interpreter/internal/lexer"
	"github.com/usuario/gaby-interpreter/internal/object"
	"github.com/usuario/gaby-interpreter/internal/parser"
	"github.com/usuario/gaby-interpreter/stdlib"
)

const GABY_EXTENSION = ".gaby"

func main() {
	// Obtener el usuario actual para saludar
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	// Inicializar el entorno global
	env := object.NewEnvironment()
	
	// Cargar las funciones de la biblioteca estándar
	stdlib.LoadStdlib(env)

	// Verificar argumentos
	args := os.Args[1:]
	if len(args) == 0 {
		// Sin argumentos, iniciar modo interactivo (REPL)
		fmt.Printf("¡Hola %s! Bienvenido al intérprete de Gaby.\n", user.Username)
		fmt.Println("Escribe 'salir()' para salir, 'ayuda()' para ver comandos disponibles.")
		startRepl(os.Stdin, os.Stdout, env)
	} else if len(args) == 1 {
		// Con un argumento, ejecutar archivo
		filename := args[0]
		
		// Verificar extensión
		if !strings.HasSuffix(filename, GABY_EXTENSION) {
			fmt.Printf("Error: El archivo debe tener extensión %s\n", GABY_EXTENSION)
			os.Exit(1)
		}
		
		executeFile(filename, env)
	} else {
		// Demasiados argumentos
		fmt.Println("Uso: gaby [archivo.gaby]")
		os.Exit(1)
	}
}

// startRepl inicia el bucle Read-Eval-Print-Loop para interacción interactiva
func startRepl(in io.Reader, out io.Writer, env *object.Environment) {
	scanner := NewLineScanner(in)
	
	for {
		fmt.Fprint(out, ">> ")
		line, more, err := scanner.Scan()
		if err != nil {
			fmt.Fprintln(out, "Error al leer entrada:", err)
			return
		}
		
		// Si necesitamos más entrada (para bloques multilinea)
		for more {
			fmt.Fprint(out, ".. ")
			nextLine, moreInput, err := scanner.Scan()
			if err != nil {
				fmt.Fprintln(out, "Error al leer entrada:", err)
				return
			}
			line += "\n" + nextLine
			more = moreInput
		}
		
		// Procesar comandos especiales del REPL
		if line == "salir()" {
			fmt.Fprintln(out, "¡Hasta luego!")
			return
		} else if line == "ayuda()" {
			printHelp(out)
			continue
		}
		
		// Evaluar la entrada
		evaluated := evaluateInput(line, env)
		
		if evaluated != nil && evaluated.Type() != object.NULL_OBJ {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

// executeFile ejecuta un archivo .gaby
func executeFile(filename string, env *object.Environment) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error al leer el archivo: %s\n", err)
		os.Exit(1)
	}
	
	// Obtener la ruta del archivo para establecer el directorio de trabajo
	absPath, err := filepath.Abs(filename)
	if err == nil {
		dir := filepath.Dir(absPath)
		os.Chdir(dir)
	}
	
	// Evaluar el contenido del archivo
	evaluated := evaluateInput(string(content), env)
	
	// Si hay un error, mostrarlo y salir
	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		fmt.Println(evaluated.Inspect())
		os.Exit(1)
	}
}

// evaluateInput evalúa una cadena de entrada y devuelve el resultado
func evaluateInput(input string, env *object.Environment) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		printParserErrors(p.Errors())
		return nil
	}
	
	return evaluator.Eval(program, env)
}

// printParserErrors imprime errores del parser
func printParserErrors(errors []string) {
	fmt.Println("¡Ops! Encontré algunos errores:")
	for _, msg := range errors {
		fmt.Printf("\t- %s\n", msg)
	}
}

// printHelp muestra los comandos disponibles
func printHelp(out io.Writer) {
	help := `
Comandos disponibles:
  salir()   - Salir del intérprete
  ayuda()   - Mostrar esta ayuda

Ejemplos básicos:
  mostrar("¡Hola mundo!")
  suma := |a, b| -> a + b
  resultado := suma(5, 3)
  mostrar(resultado)

  // Condicionales
  edad := 25
  si edad >= 18 {
    mostrar("Mayor de edad")
  } sino {
    mostrar("Menor de edad")
  }

  // Bucles
  para i desde 1 hasta 5 {
    mostrar(i)
  }

  contador := 0
  mientras contador < 5 {
    mostrar(contador)
    contador += 1
  }

  // Funciones
  fun saludar(nombre) {
    devolver "¡Hola, " + nombre + "!"
  }
  mostrar(saludar("Gaby"))

  // Clases
  clase Persona {
    texto nombre
    num edad
    
    crear(nombre, edad) {
      esto.nombre = nombre
      esto.edad = edad
    }
    
    fun presentarse() {
      mostrar("Me llamo " + esto.nombre + " y tengo " + esto.edad + " años")
    }
  }
  
  p := nuevo Persona("Juan", 30)
  p.presentarse()
`
	io.WriteString(out, help)
}

// LineScanner es un escaner que maneja múltiples líneas para bloques de código
type LineScanner struct {
	reader      io.Reader
	buffer      []byte
	position    int
	bracketOpen int
}

// NewLineScanner crea un nuevo escáner de líneas
func NewLineScanner(reader io.Reader) *LineScanner {
	return &LineScanner{
		reader:      reader,
		buffer:      make([]byte, 0, 4096),
		position:    0,
		bracketOpen: 0,
	}
}

// Scan lee una línea y determina si necesitamos más entrada
func (ls *LineScanner) Scan() (string, bool, error) {
	var buf [1]byte
	var line []byte
	
	for {
		n, err := ls.reader.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				return string(line), false, nil
			}
			return "", false, err
		}
		
		if n == 0 {
			continue
		}
		
		c := buf[0]
		if c == '\n' {
			// Verificar si estamos en medio de un bloque
			needMore := ls.bracketOpen > 0
			return string(line), needMore, nil
		}
		
		// Contar llaves abiertas/cerradas para determinar bloques multilinea
		if c == '{' {
			ls.bracketOpen++
		} else if c == '}' {
			ls.bracketOpen--
			if ls.bracketOpen < 0 {
				ls.bracketOpen = 0
			}
		}
		
		line = append(line, c)
	}
}