package main

import (
  "fmt"
  "regexp"
  "strings"
  "strconv"
  "math/rand"
  "time"
  "errors"
  "slices"
)

/* Struct representing the result(s) of a dice roll
 */
type DiceRoll struct {
  Expression string
  Results []int
}

/* Randomly rolls based on the given dice notation
 */
func rollDice(diceNotation string) (int, []int) {
  lastChar := diceNotation[len(diceNotation)-1]
  mode := "sum"
  if lastChar == '!' {
    mode = "highest"
    diceNotation = diceNotation[:len(diceNotation)-1]
  } else if lastChar == '?' {
    mode = "lowest"
    diceNotation = diceNotation[:len(diceNotation)-1]
  }
  
  parts := strings.Split(diceNotation, "d")

  numRolls, _ := strconv.Atoi(parts[0])
  sides, _ := strconv.Atoi(parts[1])

  rand.Seed(time.Now().UnixNano())

  rolls := []int{}
  result := 0
  for i := 0; i < numRolls; i++ {
    rollValue := rand.Intn(sides) + 1
    rolls = append(rolls, rollValue)

    switch mode {
    case "sum":
      result += rollValue
    case "highest":
      if rollValue > result {
        result = rollValue
      }
    case "lowest":
      if result == 0 || rollValue < result {
        result = rollValue
      }
    }
  }
  
  return result, rolls
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
  dicePattern := `\d+d\d+[!?]?`
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
func parse(tokens []string) (int, []DiceRoll, error) {
  // An array to contain the results of dice rolls
  rollResults := []DiceRoll{}

  // Regex we'll need a little further down
  valuePattern := regexp.MustCompile(`\d+d\d+|\d+`)
  operatorPattern := regexp.MustCompile(`[+\-*/]`)
  leftParenPattern := regexp.MustCompile(`\(`)
  rightParenPattern := regexp.MustCompile(`\)`)
  dicePattern := regexp.MustCompile(`\d+d\d+[!?]?`)
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
      if len(operatorStack) == 0 {
        return 0, rollResults, errors.New("Unable to parse: mismatched parens")
      }
      
      for operatorStack[len(operatorStack)-1] != "(" {
        outputQueue = append(outputQueue, operatorStack[len(operatorStack)-1])
        operatorStack = operatorStack[:len(operatorStack)-1]
      }

      if len(operatorStack) == 0 {
        return 0, rollResults, errors.New("Unable to parse: mismatched parens")
      }
      
      operatorStack = operatorStack[:len(operatorStack)-1]
    }
  }

  if slices.Contains(operatorStack, "(") {
    return 0, rollResults, errors.New("Unable to parse: mismatched parens")
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
      rollResult, rolls := rollDice(token)
      rollResults = append(rollResults, DiceRoll{
        Expression: token,
        Results: rolls,
      })
      stack = append(stack, rollResult)
    case integerPattern.MatchString(token):
      n, _ := strconv.Atoi(token)
      stack = append(stack, n)
    case operatorPattern.MatchString(token):
      if len(stack) == 0 {
        return 0, rollResults, errors.New("Unable to parse: too many operators")
      }
      
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

  if len(stack) != 1 {
    return 0, rollResults, errors.New("Unable to parse malformed input")
  }
  
  return stack[0], rollResults, nil
}

/* Parses the given expression.
 * Expression must contain only integers and dice notation,
 * and may only use the operators + - * / and ()
 */
func ParseExpression(input string) (int, []DiceRoll, error) {
  // Tokenize
  tokenized := tokenize(input)
  
  // Quick validation: Only valid tokens in input string
  inputToCompare := strings.ReplaceAll(input, " ", "")
  inputFromTokenized := strings.Join(tokenized, "")
  if inputToCompare != inputFromTokenized {
    return 0, []DiceRoll{}, errors.New("Invalid tokens found")
  }

  // Parse
  return parse(tokenized)
}

/* Given a macro and a list of values, substitutes them into
 * the macro to produce an expression with the values filled in. 
 */
func FillMacro(input string, variables[]string) string {
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

  return input
}
