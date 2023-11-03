package main

import (
  "fmt"
  "regexp"
  "strings"
  "strconv"
  "math/rand"
  "time"
)

/* Randomly rolls based on the given dice notation
 */
func rollDice(diceNotation string) int {
  parts := strings.Split(diceNotation, "d")

  numRolls, _ := strconv.Atoi(parts[0])
  sides, _ := strconv.Atoi(parts[1])

  rand.Seed(time.Now().UnixNano())

  sum := 0
  for i := 0; i < numRolls; i++ {
    sum += rand.Intn(sides) + 1
  }

  return sum
}

/* Given two operators, op1 and op2, checks if op1 has greater precedence.
 * Returns true if op1 has greater precedence (should come first)
 */
func hasGreaterPrecedence(op1 string, op2 string) bool {
  precedence := map[string]int{"+": 1, "-": 1, "*": 2, "/": 2}
  return precedence[op1] >= precedence[op2]
}

/* Convert the input string into an array of tokens.
 * Tokens can be one of three things:
 * - An integer
 * - An operator: + - / * ( )
 * - Dice notation (XdY): ex. 4d12, 3d20
 */
func tokenize(input string) []string {
  integerPattern := `\d+`
  operatorPattern := `[+\-*/()]`
  dicePattern := `\d+d\d+`
  tokenPattern := fmt.Sprintf("(%s)|(%s)|(%s)", dicePattern, integerPattern, operatorPattern)
  
  re := regexp.MustCompile(tokenPattern)
  matches := re.FindAllString(input, -1)

  return matches
}

/* Parses an array of tokens into a final value.
 * First, use shunting yard to change the array of tokens into reverse polish notation.
 * Then, use a stack procedure to evaluate the reverse polish notation,
 * rolling dice as we go!
 */
func parse(tokens []string) int {
  // Regex we'll need a little further down
  valuePattern := regexp.MustCompile(`\d+d\d+|\d+`)
  operatorPattern := regexp.MustCompile(`[+\-*/]`)
  leftParenPattern := regexp.MustCompile(`\(`)
  rightParenPattern := regexp.MustCompile(`\)`)
  dicePattern := regexp.MustCompile(`\d+d\d+`)
  integerPattern := regexp.MustCompile(`\d+`)
  
  // First, do the shunting yard algorithm to get it into reverse polish notation
  operatorStack := []string{}
  outputQueue := []string{}

  for _, token := range tokens {
    switch {
    case valuePattern.MatchString(token):
      outputQueue = append(outputQueue, token)
    case operatorPattern.MatchString(token):
      for len(operatorStack) > 0 && hasGreaterPrecedence(operatorStack[len(operatorStack)-1], token) {
        outputQueue = append(outputQueue, operatorStack[len(operatorStack)-1])
        operatorStack = operatorStack[:len(operatorStack)-1]
      }
      operatorStack = append(operatorStack, token)
    case leftParenPattern.MatchString(token):
      operatorStack = append(operatorStack, token)
    case rightParenPattern.MatchString(token):
      for operatorStack[len(operatorStack)-1] != "(" {
        outputQueue = append(outputQueue, operatorStack[len(operatorStack)-1])
        operatorStack = operatorStack[:len(operatorStack)-1]
      }
      operatorStack = operatorStack[:len(operatorStack)-1]
    }
  }

  for len(operatorStack) > 0 {
    outputQueue = append(outputQueue, operatorStack[len(operatorStack)-1])
    operatorStack = operatorStack[:len(operatorStack)-1]
  }
  
  // Now, parse the reverse polish notation
  stack := []int{}
  for _, token := range outputQueue {
    switch {
    case dicePattern.MatchString(token):
      stack = append(stack, rollDice(token))
    case integerPattern.MatchString(token):
      n, _ := strconv.Atoi(token)
      stack = append(stack, n)
    case operatorPattern.MatchString(token):
      num2 := stack[len(stack)-1]
      stack = stack[:len(stack)-1]

      num1 := stack[len(stack)-1]
      stack = stack[:len(stack)-1]

      switch token {
      case "+":
        stack = append(stack, num1 + num2)
      case "-":
        stack = append(stack, num1 - num2)
      case "*":
        stack = append(stack, num1 * num2)
      case "/":
        stack = append(stack, num1 / num2)
      }
    }
  }

  return stack[0]
}

/* Parses the given expression.
 * Expression must contain only integers and dice notation,
 * and may only use the operators + - * / and ()
 */
func ParseExpression(input string) int {
  return parse(tokenize(input))
}

/* Parses the given macro. Macros are expressions which contain letters (A, B, C..)
 * representing variables in the expression. The variables argument contains the values
 * used to substitute in the macro.
 */
func ParseMacro(input string, variables []string) int {
  // Set up variables for macros (A, B, C, etc..)
  variablesMap := make(map[string]string)
  for i, v := range variables {
    if i < 26 {
      key := string('A' + i)
      variablesMap[key] = v
    }
  }

  // Replace variables in the input string
  for key, value := range variablesMap {
    input = strings.Replace(input, key, value, -1)
  }

  // Parse
  return parse(tokenize(input))
}
