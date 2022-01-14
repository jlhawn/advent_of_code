package main

import (
	"bufio"
	// "constraints"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	// "time"

	// "../../slices"
	"../../streams"
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

func parseRule(line string) Rule {
	id, line, ok := strings.Cut(line, ": ")
	if !ok { panic("unable to cut line") }

	if rule := parseStringRule(id, line); rule != nil {
		return rule
	}

	return parseOptionRule(id, line)
}

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Size() int {
	return len(s.items)
}

func (s *Stack[T]) IsEmpty() bool {
	return s.Size() == 0
}

func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (item T, ok bool) {
	if len(s.items) == 0 {
		return item, false
	}

	end := len(s.items)-1
	item = s.items[end]
	s.items = s.items[:end]
	return item, true
}

func (s *Stack[T]) Peek() (item T, ok bool) {
	if len(s.items) == 0 {
		return item, false
	}

	end := len(s.items)-1
	item = s.items[end]
	return item, true
}

type Lexer struct {
	input string
	cursorPos int
}

func NewLexer(s string) *Lexer {
	return &Lexer{input: s}
}

func (l *Lexer) CursorPos() int {
	return l.cursorPos
}

func (l *Lexer) String() string {
	return l.input[l.cursorPos:]
}

func (l *Lexer) Expect(s string, expectDone bool) (ok bool) {
	startingPos := l.cursorPos
	defer func() {
		if !ok {
			l.Reset(startingPos)
		}
	}()
	// Ignore leading spaces.
	for l.input[l.cursorPos] == ' ' {
		l.cursorPos++
	}
	if !strings.HasPrefix(l.String(), s) {
		return false
	}
	l.cursorPos += len(s)
	return expectDone == l.Done()
}

func (l *Lexer) Reset(cursorPos int) {
	l.cursorPos = cursorPos
}

func (l *Lexer) Done() bool {
	return l.cursorPos == len(l.input)
}

type RuleSet map[string]Rule

type ParseTree struct {
	RuleTag string
	Value string
	SubTrees []*ParseTree
}

func NewParseTree(ruleTag, value string, subTrees ...*ParseTree) *ParseTree {
	return &ParseTree{ruleTag, value, subTrees}
}

func (t *ParseTree) BuildString(b *strings.Builder) {
	b.WriteString("(")
	b.WriteString(t.RuleTag)
	if t.Value != "" {
		b.WriteString(" ")
		b.WriteString(t.Value)
	}
	for _, subTree := range t.SubTrees {
		b.WriteString(" ")
		subTree.BuildString(b)
	}
	b.WriteString(")")
}

func (t *ParseTree) String() string {
	var b strings.Builder
	t.BuildString(&b)
	return b.String()
}

type Tracer interface {
	Next(tag string) Tracer
	Trace(message string)
}

type noTrace struct{}

func (n noTrace) Next(tag string) Tracer { return n }

func (n noTrace) Trace(message string) { return }

type stringTracer string

func NewStringTracer(tag string) Tracer {
	return stringTracer(fmt.Sprintf("%s -> ", tag))
}

func (s stringTracer) Next(tag string) Tracer {
	return stringTracer(fmt.Sprintf("%s%s -> ", s, tag))
}

func (s stringTracer) Trace(message string) {
	fmt.Printf("%s%s\n", s, message)
}

type ParseTreeStream = streams.Stream[*ParseTree]

type Rule interface {
	ID() string
	Matches(t Tracer, l *Lexer, expectDone bool) ParseTreeStream
}

type FinalizerRule interface {
	Rule
	Finalize(rulesByID RuleSet)
}

type stringRule struct {
	id, val string
}

func newStringRule(id, val string) Rule {
	return &stringRule{id, val}
}

func parseStringRule(id, line string) Rule {
	if !(strings.HasPrefix(line, "\"") && strings.HasSuffix(line, "\"")) {
		return nil
	}
	val := strings.TrimSuffix(strings.TrimPrefix(line, "\""), "\"")
	return newStringRule(id, val)
}

func (s *stringRule) ID() string {
	return s.id
}

func (s *stringRule) Matches(t Tracer, l *Lexer, expectDone bool) ParseTreeStream {
	productions := streams.Once(func() (item *ParseTree, ok bool) {
		t.Trace("next")
		if l.Expect(s.val, expectDone) {
			t.Trace(fmt.Sprintf("%q", s.val))
			return NewParseTree(s.ID(), strconv.Quote(s.val)), true
		}
		return nil, false
	})

	productions = streams.WithInitializer(productions, func() { t.Trace("begin") })

	cursorStartPos := l.CursorPos()
	return streams.WithFinalizer(productions, func() {
		t.Trace("end")
		// We get to this point when our parent rule is asking for our next
		// match but we don't have any. Clearly we should set the cursor back
		// to the position it was at when we started.
		l.Reset(cursorStartPos)
	})
}

type multiRule struct {
	id string
	// A multiRule is always either a direct rule or an option of a rule.
	// If it's a direct rule, id and tag will be the same.
	// If it's an option of a rule, id will be the outer rule id and
	// tag will be the outer rule id with a subscript for which
	// option index it is, e.g., expr[0].
	tag string
	ruleIDs []string
	rules []Rule
}

func parseMultiRule(id, line string) Rule {
	return newMultiRule(id, strings.Split(line, " "))
}

func newMultiRule(id string, ruleIDs []string) Rule {
	return &multiRule{
		id: id,
		tag: id,
		ruleIDs: ruleIDs,
		rules: make([]Rule, len(ruleIDs)),
	}
}

func (m *multiRule) Finalize(rulesByID RuleSet) {
	for i, ruleID := range m.ruleIDs {
		m.rules[i] = rulesByID[ruleID]
	}
}

func (m *multiRule) ID() string {
	return m.id
}

func (m *multiRule) tracerTag(target int) string {
	var b strings.Builder
	for i := 0; i < len(m.rules); i++ {
		if i > 0 {
			b.WriteString(" ")
		}
		if target == i {
			b.WriteString("[")
		}
		b.WriteString(m.rules[i].ID())
		if target == i {
			b.WriteString("]")
		}
	}
	return b.String()
}

func (m *multiRule) Matches(t Tracer, l *Lexer, expectDone bool) ParseTreeStream {
	// Each of the sub-rules will return a stream, they'll need to be merged together
	// somehow.
	// Let's say there were 2 sub-rules. For each match in the stream output for the
	// first rule we'll need to combine it with the a production from the stream of
	// output from the second rule in order to make a production for this rule.
	// If there's a 3rd rule, each of its productions will have to be combined with
	// the production from the first and second stream to make a production for this
	// rule.
	// First we'll get the stream of productions from the first rule. For each
	// production from the first rule we'll get a stream of producitons from the
	// second rule. For each production from the second rule, we'll get a stream of
	// productions from the third rule. Do this until we get to the end of the list
	// of sub-rules. In this case, we can produce the parse tree which combines the
	// productions of each of the rule's production streams. If a stream ends, go
	// back to the stream for the previous rule and get the next production in its
	// stream and so on ... until we've gone through the stream of productions for
	// the first rule.
	// So we need to create a slice of streams with one entry for each sub-rule.
	var subRuleStreams Stack[ParseTreeStream]
	subRuleProductions := make([]*ParseTree, len(m.rules))

	productions := streams.Generator(func() (item *ParseTree, ok bool) {
		t.Trace("next")
		// Only true the first time the generator is called.
		if subRuleStreams.IsEmpty() {
			isEnd := len(m.rules) == 1
			subRuleStreams.Push(m.rules[0].Matches(t.Next(m.tracerTag(0)), l, isEnd && expectDone))
		}

		for !subRuleStreams.IsEmpty() {
			// - Peek at the top stream on the stack.
			topStream, _ := subRuleStreams.Peek()
			// - Get the next production from that stream.
			item, ok := topStream.Next()
			// - If the stream is exhausted, pop off the top item and
			//   continue to the next iteration.
			if !ok {
				subRuleStreams.Pop()
				continue
			}
			// - Add the production to the sub-rule matches slice.
			stackSize := subRuleStreams.Size()
			subRuleProductions[stackSize-1] = item
			// - If we're not on the last sub-rule push
			//   the next sub-rules matches onto the stack
			//   and continue.
			if stackSize < len(m.rules) {
				nextIsEnd := stackSize == len(m.rules)-1
				subRuleStreams.Push(m.rules[stackSize].Matches(t.Next(m.tracerTag(stackSize)), l, nextIsEnd && expectDone))
				continue
			}
			// - If we are on the last sub-rule, output a
			//   production of our own which contains all of the
			//   sub-rule matches.
			subTrees := make([]*ParseTree, len(m.rules))
			for i, subRuleProduction := range subRuleProductions {
				subTrees[i] = subRuleProduction
			}
			return NewParseTree(m.tag, "", subTrees...), true
		}

		return nil, false
	})

	productions = streams.WithInitializer(productions, func() { t.Trace("begin") })
	return streams.WithFinalizer(productions, func() { t.Trace("end") })
}

type optionsRule struct {
	id string
	options []Rule
}

func newOptionsRule(id string, options []Rule) Rule {
	return &optionsRule{id: id, options: options}
}

func parseOptionRule(id, line string) Rule {
	parts := strings.Split(line, " | ")

	if len(parts) == 1 {
		return parseMultiRule(id, parts[0])
	}

	options := make([]Rule, len(parts))
	for i, part := range parts {
		options[i] = parseOptionalMultiRule(id, i, part)
	}

	return newOptionsRule(id, options)
}

func parseOptionalMultiRule(id string, optionIndex int, line string) Rule {
	return newOptionalMultiRule(id, optionIndex, strings.Split(line, " "))
}

func newOptionalMultiRule(id string, optionIndex int, ruleIDs []string) Rule {
	return &multiRule{
		id: id,
		tag: fmt.Sprintf("%s[%d]", id, optionIndex),
		ruleIDs: ruleIDs,
		rules: make([]Rule, len(ruleIDs)),
	}
}

func (m *optionsRule) Finalize(rulesByID RuleSet) {
	for _, rule := range m.options {
		rule.(FinalizerRule).Finalize(rulesByID)
	}
}

func (m *optionsRule) ID() string {
	return m.id
}

func (m *optionsRule) Matches(t Tracer, l *Lexer, expectDone bool) ParseTreeStream {
	// For this type of rule, we want to produce a stream which returns
	// productions from each of our options in order. The simple way to do
	// this is with a flat map over the options.
	optionIndex := 0
	productions := streams.FlatMap(
		streams.Map(
			streams.FromItems(m.options...),
			func(rule Rule) Rule {
				t.Trace("next")
				return rule
			},
		),
		func(rule Rule) ParseTreeStream {
			subTracer := t.Next(fmt.Sprintf("%s[%d]", m.ID(), optionIndex))
			optionIndex++
			return rule.Matches(subTracer, l, expectDone)
		},
	)
	productions = streams.WithInitializer(productions, func() { t.Trace("begin") })
	return streams.WithFinalizer(productions, func() { t.Trace("end") })
}

func loadRulesAndMessages() (rulesByID RuleSet, messages []string) {
	rulesByID = RuleSet{}
	lines := readInputLines()
	for i, line := range lines {
		if len(line) == 0 {
			messages = lines[i+1:]
			break
		}
		rule := parseRule(line)
		rulesByID[rule.ID()] = rule
	}

	finalizeRules(rulesByID)

	return rulesByID, messages
}

func finalizeRules(rulesByID RuleSet) {
	for _, rule := range rulesByID {
		if r, ok := rule.(FinalizerRule); ok {
			r.Finalize(rulesByID)
		}
	}
}

func main() {
	rulesByID, messages := loadRulesAndMessages()

	rule0, ok := rulesByID["0"]
	if !ok { panic("no rule 0") }

	var validMessages []string
	for _, message := range messages {
		lexer := NewLexer(message)
		productions := rule0.Matches(noTrace{}, lexer, true)
		if _, ok := productions.Next(); ok {
			validMessages = append(validMessages, message)
		}
	}

	fmt.Printf("Part 1: found %d valid messages\n", len(validMessages))

	rulesByID["8"] = &optionsRule{
		id: "8",
		options: []Rule{
			parseOptionalMultiRule("8", 0, "42"),
			parseOptionalMultiRule("8", 1, "42 8"),
		},
	}
	rulesByID["11"] = &optionsRule{
		id: "11",
		options: []Rule{
			parseOptionalMultiRule("11", 0, "42 31"),
			parseOptionalMultiRule("11", 1, "42 11 31"),
		},
	}
	finalizeRules(rulesByID)

	rule0, _ = rulesByID["0"]
	validMessages = nil
	for _, message := range messages {
		lexer := NewLexer(message)
		// productions := rule0.Matches(NewStringTracer("0"), lexer, true)
		productions := rule0.Matches(noTrace{}, lexer, true)
		if _, ok := productions.Next(); ok {
			validMessages = append(validMessages, message)
		}
	}

	fmt.Printf("Part 2: found %d valid messages\n", len(validMessages))
}
