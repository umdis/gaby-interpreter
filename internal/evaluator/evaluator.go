package evaluator

import (
	"fmt"
	"github.com/umdis/gaby-interpreter/internal/object"
	"github.com/umdis/gaby-interpreter/internal/parser"
)

// Objetos singleton para optimizar la creación de objetos comunes
var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

// Eval evalúa un nodo del AST y devuelve un objeto
func Eval(node parser.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	// Sentencias
	case *parser.Program:
		return evalProgram(node, env)
	case *parser.ExpressionStatement:
		return Eval(node.Expression, env)
	case *parser.BlockStatement:
		return evalBlockStatement(node, env)
	case *parser.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
		return val
	case *parser.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	// Expresiones
	case *parser.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *parser.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *parser.StringLiteral:
		return &object.String{Value: node.Value}
	case *parser.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *parser.NullLiteral:
		return NULL
	case *parser.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *parser.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *parser.IfExpression:
		return evalIfExpression(node, env)
	case *parser.WhileExpression:
		return evalWhileExpression(node, env)
	case *parser.ForExpression:
		return evalForExpression(node, env)
	case *parser.Identifier:
		return evalIdentifier(node, env)
	case *parser.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{
			Parameters: params,
			Body:       body,
			Env:        env,
			Name:       node.Name,
		}
	case *parser.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args)
	case *parser.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *parser.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	case *parser.HashLiteral:
		return evalHashLiteral(node, env)
	case *parser.DotExpression:
		obj := Eval(node.Object, env)
		if isError(obj) {
			return obj
		}
		return evalDotExpression(obj, node.Property.Value)
	case *parser.ClassLiteral:
		return evalClassLiteral(node, env)
	case *parser.NewExpression:
		return evalNewExpression(node, env)
	}

	return NULL
}

func evalProgram(program *parser.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *parser.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil && (result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ) {
			return result
		}
	}

	return result
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("operador de prefijo desconocido: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	switch right.Type() {
	case object.INTEGER_OBJ:
		value := right.(*object.Integer).Value
		return &object.Integer{Value: -value}
	case object.FLOAT_OBJ:
		value := right.(*object.Float).Value
		return &object.Float{Value: -value}
	default:
		return newError("operador de prefijo desconocido: -%s", right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.FLOAT_OBJ:
		// Convertir entero a float
		intValue := left.(*object.Integer).Value
		floatValue := float64(intValue)
		leftAsFloat := &object.Float{Value: floatValue}
		return evalFloatInfixExpression(operator, leftAsFloat, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.INTEGER_OBJ:
		// Convertir entero a float
		intValue := right.(*object.Integer).Value
		floatValue := float64(intValue)
		rightAsFloat := &object.Float{Value: floatValue}
		return evalFloatInfixExpression(operator, left, rightAsFloat)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case operator == "y":
		return evalLogicalAndOperator(left, right)
	case operator == "o":
		return evalLogicalOrOperator(left, right)
	case left.Type() != right.Type():
		return newError("tipo de operando no válido: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("operador desconocido: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("división por cero")
		}
		return &object.Integer{Value: leftVal / rightVal}
	case "%":
		if rightVal == 0 {
			return newError("módulo por cero")
		}
		return &object.Integer{Value: leftVal % rightVal}
	case "^":
		// Implementación simple de potencia para enteros
		result := int64(1)
		for i := int64(0); i < rightVal; i++ {
			result *= leftVal
		}
		return &object.Integer{Value: result}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("operador desconocido: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.Float).Value

	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("división por cero")
		}
		return &object.Float{Value: leftVal / rightVal}
	case "%":
		if rightVal == 0 {
			return newError("módulo por cero")
		}
		// Implementación del módulo para floats
		return &object.Float{Value: float64(int64(leftVal) % int64(rightVal))}
	case "^":
		// Implementación simple de potencia para floats
		result := 1.0
		for i := 0.0; i < rightVal; i++ {
			result *= leftVal
		}
		return &object.Float{Value: result}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("operador desconocido: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch operator {
	case "+":
		return &object.String{Value: leftVal + rightVal}
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("operador desconocido: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalLogicalAndOperator(left, right object.Object) object.Object {
	if !isTruthy(left) {
		return left
	}
	return right
}

func evalLogicalOrOperator(left, right object.Object) object.Object {
	if isTruthy(left) {
		return left
	}
	return right
}

func evalIfExpression(ie *parser.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func evalWhileExpression(we *parser.WhileExpression, env *object.Environment) object.Object {
	var result object.Object = NULL

	for {
		condition := Eval(we.Condition, env)
		if isError(condition) {
			return condition
		}

		if !isTruthy(condition) {
			break
		}

		result = Eval(we.Body, env)
		if isError(result) {
			return result
		}

		// Manejar sentencias de retorno, pero no salir del bucle por ellas
		if result != nil && result.Type() == object.RETURN_VALUE_OBJ {
			return result
		}
	}

	return result
}

func evalForExpression(fe *parser.ForExpression, env *object.Environment) object.Object {
	// Crear un entorno separado para el bucle
	loopEnv := object.NewEnclosedEnvironment(env)

	// Inicialización
	if fe.Init != nil {
		initResult := Eval(fe.Init, loopEnv)
		if isError(initResult) {
			return initResult
		}
	}

	var result object.Object = NULL

	for {
		// Condición
		if fe.Condition != nil {
			condition := Eval(fe.Condition, loopEnv)
			if isError(condition) {
				return condition
			}
			if !isTruthy(condition) {
				break
			}
		}

		// Cuerpo
		result = Eval(fe.Body, loopEnv)
		if isError(result) {
			return result
		}

		// Manejar sentencias de retorno, pero no salir del bucle por ellas
		if result != nil && result.Type() == object.RETURN_VALUE_OBJ {
			return result
		}

		// Actualización
		if fe.Update != nil {
			updateResult := Eval(fe.Update, loopEnv)
			if isError(updateResult) {
				return updateResult
			}
		}
	}

	return result
}

func evalIdentifier(node *parser.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identificador no encontrado: " + node.Value)
}

func evalExpressions(exps []parser.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("no es una función: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		if paramIdx < len(args) {
			env.Set(param.Value, args[paramIdx])
		} else {
			// Si no hay suficientes argumentos, asignar NULL
			env.Set(param.Value, NULL)
		}
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("operador de índice no soportado: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}

func evalHashLiteral(node *parser.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("clave no utilizable como hash: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("clave no utilizable como hash: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func evalDotExpression(obj object.Object, property string) object.Object {
	switch obj := obj.(type) {
	case *object.Instance:
		// Buscar propiedad en la instancia
		if val, ok := obj.Properties[property]; ok {
			return val
		}

		// Buscar método en la clase
		if method, ok := obj.Class.Methods[property]; ok {
			// Enlazar el método a esta instancia (this/esto)
			boundMethod := &object.Function{
				Parameters: method.Parameters,
				Body:       method.Body,
				Env:        method.Env,
				Name:       method.Name,
			}

			// Crear un entorno para el método con 'esto' configurado
			methodEnv := object.NewEnclosedEnvironment(obj.Env)
			methodEnv.Set("esto", obj)
			boundMethod.Env = methodEnv

			return boundMethod
		}

		return newError("propiedad o método no encontrado: %s", property)
	case *object.String:
		// Añadir métodos incorporados para strings
		switch property {
		case "longitud":
			return &object.Integer{Value: int64(len(obj.Value))}
		// Añadir más métodos de string según sea necesario
		}
		return newError("propiedad no encontrada en string: %s", property)
	case *object.Array:
		// Añadir métodos incorporados para arrays
		switch property {
		case "longitud":
			return &object.Integer{Value: int64(len(obj.Elements))}
		// Añadir más métodos de array según sea necesario
		}
		return newError("propiedad no encontrada en array: %s", property)
	default:
		return newError("acceso a propiedad no soportado para: %s", obj.Type())
	}
}

func evalClassLiteral(node *parser.ClassLiteral, env *object.Environment) object.Object {
	class := &object.Class{
		Name:       node.Name.Value,
		Properties: make(map[string]object.Object),
		Methods:    make(map[string]*object.Function),
	}

	// Procesar propiedades
	for _, propNode := range node.Properties {
		propValue := Eval(propNode.Value, env)
		if isError(propValue) {
			return propValue
		}
		class.Properties[propNode.Name.Value] = propValue
	}

	// Procesar métodos
	for _, methodNode := range node.Methods {
		methodEnv := object.NewEnclosedEnvironment(env)
		method := &object.Function{
			Parameters: methodNode.Parameters,
			Body:       methodNode.Body,
			Env:        methodEnv,
			Name:       methodNode.Name,
		}
		class.Methods[methodNode.Name] = method
	}

	// Almacenar la clase en el entorno
	env.Set(node.Name.Value, class)

	return class
}

func evalNewExpression(node *parser.NewExpression, env *object.Environment) object.Object {
	classObj := Eval(node.Class, env)
	if isError(classObj) {
		return classObj
	}

	class, ok := classObj.(*object.Class)
	if !ok {
		return newError("no es una clase: %s", classObj.Type())
	}

	// Crear un nuevo entorno para la instancia
	instanceEnv := object.NewEnclosedEnvironment(env)

	// Crear la instancia
	instance := &object.Instance{
		Class:      class,
		Properties: make(map[string]object.Object),
		Env:        instanceEnv,
	}

	// Configurar 'esto' para referir a la instancia
	instanceEnv.Set("esto", instance)

	// Copiar propiedades de la clase a la instancia
	for name, value := range class.Properties {
		instance.Properties[name] = value
	}

	// Llamar al constructor si existe
	if constructor, ok := class.Methods["crear"]; ok {
		// Preparar los argumentos
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		// Configurar el entorno del constructor
		constructorEnv := object.NewEnclosedEnvironment(constructor.Env)
		constructorEnv.Set("esto", instance)

		// Configurar los parámetros del constructor
		for paramIdx, param := range constructor.Parameters {
			if paramIdx < len(args) {
				constructorEnv.Set(param.Value, args[paramIdx])
			}
		}

		// Ejecutar el constructor
		Eval(constructor.Body, constructorEnv)
	}

	return instance
}

// Funciones auxiliares

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		switch obj.Type() {
		case object.INTEGER_OBJ:
			return obj.(*object.Integer).Value != 0
		case object.FLOAT_OBJ:
			return obj.(*object.Float).Value != 0
		case object.STRING_OBJ:
			return obj.(*object.String).Value != ""
		default:
			return true
		}
	}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// Funciones incorporadas (builtins)
var builtins = map[string]*object.Builtin{
	"longitud": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("número incorrecto de argumentos para 'longitud': se esperaba 1, se obtuvo %d", len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			default:
				return newError("argumento para 'longitud' no soportado, se obtuvo %s", args[0].Type())
			}
		},
	},
	"mostrar": {
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
	// Añadir más funciones incorporadas según sea necesario
}