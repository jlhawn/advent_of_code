package main

import (
	"bufio"
	// "constraints"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	// "time"

	"../../slices"
	// "../../streams"
)

const inputFilename = "./INPUT"

func readInputLines() []string {
	inputFile, err := os.Open(inputFilename)
	if err != nil {
		panic(err)
	}
	defer inputFile.Close()

	var (
		line string
		lineNum int
		lines []string
	)

	bufReader := bufio.NewReader(inputFile)
	for {
		lineNum++
		line, err = bufReader.ReadString('\n')
		if err != nil {
			err = fmt.Errorf("unable to read line %d: %w", lineNum, err)
			break
		}
		lines = append(lines, strings.TrimSpace(line))
	}
	if !errors.Is(err, io.EOF) {
		panic(err)
	}

	return lines
}

type TokenTypeID int

const (
	TokenTypeIDInteger TokenTypeID = iota
	TokenTypeIDPlus
	TokenTypeIDMult
	TokenTypeIDOpenParen
	TokenTypeIDCloseParen
	TokenTypeIDEnd
)

func (t TokenTypeID) String() string {
	switch t {
	case TokenTypeIDInteger:
		return "INTEGER"
	case TokenTypeIDPlus:
		return "PLUS"
	case TokenTypeIDMult:
		return "MULTIPLY"
	case TokenTypeIDOpenParen:
		return "OPEN_PAREN"
	case TokenTypeIDCloseParen:
		return "CLOSE_PAREN"
	case TokenTypeIDEnd:
		return "END_OF_LINE"
	default:
		return "<UNKOWN_TOKEN_TYPE>"
	}
}

var TokenPatterns = map[TokenTypeID]*regexp.Regexp{
	TokenTypeIDInteger:    regexp.MustCompile(`^[0-9]+`),
	TokenTypeIDPlus:       regexp.MustCompile(`^\+`),
	TokenTypeIDMult:       regexp.MustCompile(`^\*`),
	TokenTypeIDOpenParen:  regexp.MustCompile(`^\(`),
	TokenTypeIDCloseParen: regexp.MustCompile(`^\)`),
}

type Token struct {
	TokenTypeID
	Raw string
}

func (t Token) String() string {
	return fmt.Sprintf("%s %q", t.TokenTypeID, t.Raw)
}

type TokenSequence struct {
	tokens []*Token
	index int
}

func (t *TokenSequence) Peek() *Token {
	if t.index < len(t.tokens) {
		return t.tokens[t.index]
	}
	return nil
}

var ErrNoNextToken = errors.New("no next token")

func (t *TokenSequence) Next() *Token {
	token := t.Peek()
	if token != nil {
		t.index++
	}
	return token
}

func (t *TokenSequence) PeekAny(typeIDs ...TokenTypeID) bool {
	token := t.Peek()
	if token == nil {
		return false
	}

	return slices.Any(func(tokenID TokenTypeID) bool { return token.TokenTypeID == tokenID }, typeIDs...)
}

func (t *TokenSequence) ExpectNone() error {
	_, err := t.expectAny(false)
	return err
}

func (t *TokenSequence) Expect(typeID TokenTypeID) error {
	_, err := t.expectAny(false, typeID)
	return err
}

func (t *TokenSequence) ExpectNext(typeID TokenTypeID) (*Token, error) {
	return t.expectAny(true, typeID)
}

func (t *TokenSequence) ExpectAny(typeIDs ...TokenTypeID) error {
	_, err := t.expectAny(false, typeIDs...)
	return err
}

func (t *TokenSequence) ExpectNextAny(typeIDs ...TokenTypeID) (*Token, error) {
	return t.expectAny(true, typeIDs...)
}

func (t *TokenSequence) expectAny(next bool, typeIDs ...TokenTypeID) (*Token, error) {
	var token *Token
	if next {
		token = t.Next()
	} else {
		token = t.Peek()
	}

	if len(typeIDs) == 0 {
		if token != nil {
			return token, fmt.Errorf("expected no remaining tokens but got %s", token)
		}
		return token, nil
	}
	
	if token == nil {
		if len(typeIDs) == 1 {
			return token, fmt.Errorf("expected %s token but got %w", typeIDs[0], ErrNoNextToken)
		}
		return token, fmt.Errorf("expected any of %s tokens but got %w", typeIDs, ErrNoNextToken)
	}

	if !slices.Any(func(typeID TokenTypeID) bool { return token.TokenTypeID == typeID }, typeIDs...) {
		if len(typeIDs) == 1 {
			return token, fmt.Errorf("expected %s token but got %s", typeIDs[0], token.TokenTypeID)
		}
		return token, fmt.Errorf("expected any of %s tokens but got %s", typeIDs, token)
	}

	return token, nil
}

func tokenize(line string) *TokenSequence {
	var tokens []*Token
	line = strings.TrimSpace(line)
	for len(line) > 0 {
		var token *Token
		for tokenType, pattern := range TokenPatterns {
			loc := pattern.FindStringIndex(line)
			if loc == nil {
				continue
			}
			token = &Token{tokenType, line[loc[0]:loc[1]]}
			line = line[loc[1]:]
			break
		}
		if token == nil {
			panic(fmt.Errorf("unable to find token in remaning line: %q", line))
		}
		tokens = append(tokens, token)
		line = strings.TrimSpace(line)
	}
	tokens = append(tokens, &Token{TokenTypeID: TokenTypeIDEnd})
	return &TokenSequence{tokens: tokens}
}

type Expr interface {
	Value() int
}

type Integer int

func (i Integer) Value() int {
	return int(i)
}

type BinaryOp struct {
	Op func(a, b Expr) int
	Left, Right Expr
}

func Add(a, b Expr) int {
	return a.Value() + b.Value()
}

func Multiply(a, b Expr) int {
	return a.Value() * b.Value()
}

func (b BinaryOp) Value() int {
	return b.Op(b.Left, b.Right)
}

func AddOp(a, b Expr) Expr {
	return BinaryOp{Add, a, b}
}

func MultiplyOp(a, b Expr) Expr {
	return BinaryOp{Multiply, a, b}
}

/*

PART 1 Grammar

int = '0' - '9'

atom = int
     | '(' expr ')'

expr = atom
     | expr '+' atom
     | expr '*' atom

*/

func parseExpression1(t *TokenSequence) (Expr, error) {
	left, err := parseAtom(t, parseExpression1)
	if err != nil {
		return nil, err
	}

	for !t.PeekAny(TokenTypeIDCloseParen, TokenTypeIDEnd) {
		op, err := t.ExpectNextAny(TokenTypeIDPlus, TokenTypeIDMult)
		if err != nil {
			return nil, err
		}

		right, err := parseAtom(t, parseExpression1)
		if err != nil {
			return nil, err
		}

		switch op.TokenTypeID {
		case TokenTypeIDPlus:
			left = AddOp(left, right)
		case TokenTypeIDMult:
			left = MultiplyOp(left, right)
		default:
			panic("unexpected token type id")
		}
	}

	return left, nil
}

func parseAtom(t *TokenSequence, parseExpression func(*TokenSequence) (Expr, error)) (Expr, error) {
	if t.PeekAny(TokenTypeIDInteger) {
		return parseInteger(t)
	}

	if _, err := t.ExpectNext(TokenTypeIDOpenParen); err != nil {
		return nil, err
	}

	expr, err := parseExpression(t)
	if err != nil {
		return nil, err
	}

	if _, err := t.ExpectNext(TokenTypeIDCloseParen); err != nil {
		return nil, err
	}

	return expr, nil
}

func parseInteger(t *TokenSequence) (Expr, error) {
	token, err := t.ExpectNext(TokenTypeIDInteger)
	if err != nil {
		return nil, err
	}

	val, err := strconv.Atoi(token.Raw)
	if err != nil {
		return nil, fmt.Errorf("unable to parse integer token %q: %w", token.Raw, err)
	}

	return Integer(val), nil
}

/*

PART 2 Grammar

int = '0' - '9'

atom = int
     | '(' expr ')'

sum = atom
    | sum '+' atom

expr = sum
     | expr '*' sum

*/

func parseSum2(t *TokenSequence) (Expr, error) {
	left, err := parseAtom(t, parseExpression2)
	if err != nil {
		return nil, err
	}

	for !t.PeekAny(TokenTypeIDMult, TokenTypeIDCloseParen, TokenTypeIDEnd) {
		if _, err := t.ExpectNext(TokenTypeIDPlus); err != nil {
			return nil, err
		}

		right, err := parseAtom(t, parseExpression2)
		if err != nil {
			return nil, err
		}

		left = AddOp(left, right)
	}

	return left, nil
}

func parseExpression2(t *TokenSequence) (Expr, error) {
	left, err := parseSum2(t)
	if err != nil {
		return nil, err
	}

	for !t.PeekAny(TokenTypeIDCloseParen, TokenTypeIDEnd) {
		if _, err := t.ExpectNext(TokenTypeIDMult); err != nil {
			return nil, err
		}

		right, err := parseSum2(t)
		if err != nil {
			return nil, err
		}

		left = MultiplyOp(left, right)
	}

	return left, nil
}

func main() {
	var sum1, sum2 int
	for i, line := range readInputLines() {
		tokens := tokenize(line)
		expr1, err := parseExpression1(tokens)
		if err != nil {
			panic(fmt.Errorf("line %d: %w", i+1, err))
		}

		tokens = tokenize(line)
		expr2, err := parseExpression2(tokens)
		if err != nil {
			panic(fmt.Errorf("line %d: %w", i+1, err))
		}

		sum1 += expr1.Value()
		sum2 += expr2.Value()
	}
	fmt.Printf("Part 1 Sum of expressions: %d\n", sum1)
	fmt.Printf("Part 2 Sum of expressions: %d\n", sum2)
}
