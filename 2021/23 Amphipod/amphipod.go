package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"../../streams"
)

type SequenceNode[T any] struct {
	Item T
	Next *SequenceNode[T]
}

type Sequence[T any] struct {
	First *SequenceNode[T]
	Size int
}

func (s *Sequence[T]) ShallowCopy() *Sequence[T] {
	return &Sequence[T]{
		First: s.First,
		Size: s.Size,
	}
}

type sequenceStream[T any] struct {
	node *SequenceNode[T]
}

func (s *sequenceStream[T]) Next() (item T, ok bool) {
	if s.node == nil {
		return item, false
	}
	item = s.node.Item
	s.node = s.node.Next
	return item, true
}

func (s *Sequence[T]) Stream() streams.Stream[T] {
	return &sequenceStream[T]{node: s.First}
}

func (s *Sequence[T]) AsSlice() []T {
	if s == nil {
		return nil
	}
	items := make([]T, s.Size)
	node := s.First
	for i := 0; node != nil; i++ {
		items[i] = node.Item
		node = node.Next
	}
	return items
}

func (s *Sequence[T]) Prepend(val T) {
	s.Size++
	node := &SequenceNode[T]{Item: val}

	if s.First == nil {
		s.First = node
		return
	}

	node.Next = s.First
	s.First = node
}

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

func loadAmphipodTypes() []AmphipodType {
	rawLines := readInputLines()
	// First line is all '#'
	// Second line is an empty hallway.
	rawLines = rawLines[2:]
	
	// Every other line is a row of room positions.
	var amphipodTypes []AmphipodType
	for _, line := range rawLines {
		line = strings.Trim(line, " #")
		if len(line) == 0 {
			break // End of room rows.
		}

		parts := strings.Split(line, "#")
		if len(parts) != 4 { panic("expected 4 amphipod types") }

		for _, part := range parts {
			switch part {
			case "A":
				amphipodTypes = append(amphipodTypes, Amber)
			case "B":
				amphipodTypes = append(amphipodTypes, Bronze)
			case "C":
				amphipodTypes = append(amphipodTypes, Copper)
			case "D":
				amphipodTypes = append(amphipodTypes, Desert)
			default:
				panic(fmt.Errorf("unexpected amphipod type: %q", part))
			}
		}
	}
	return amphipodTypes
}

type AmphipodType int

const (
	Amber AmphipodType = iota
	Bronze
	Copper
	Desert
)

func (t AmphipodType) String() string {
	switch t {
	case Amber:
		return "A"
	case Bronze:
		return "B"
	case Copper:
		return "C"
	case Desert:
		return "D"
	default:
		return "?"
	}
}

var MoveCosts map[AmphipodType]int

func init() {
	MoveCosts = map[AmphipodType]int {
		Amber:  1,
		Bronze: 10,
		Copper: 100,
		Desert: 1000,
	}
}

type Space struct {
	Left, Right, Up, Down *Space

	Occupant *Amphipod

	// Whether this space is a room space.
	IsRoom bool
	// Target Amphipod type if IsRoom == true.
	TargetType AmphipodType
}

func (s *Space) String() string {
	if s.Occupant != nil {
		return s.Occupant.String()
	}
	return "."
}

func (s *Space) ID() string {
	index := 0
	at := s
	for at.IsRoom && at.Up.IsRoom {
		// Index is how far this room space is from the hallway.
		at = at.Up
		index++
	}

	if at.IsRoom {
		// Index is how far this room space is from the hallway.
		return fmt.Sprintf("Room %s-%d", at.TargetType, index)
	}

	// Index is how far we are from the left end of the hallway.
	for at.Left != nil {
		index++
		at = at.Left
	}
	return fmt.Sprintf("Hallway %d", index)
}

type Amphipod struct {
	Type AmphipodType
	Location *Space
}

func (a *Amphipod) String() string {
	return a.Type.String()
}

func (a *Amphipod) IsDone() bool {
	// Must at least be in a room of the correct type.
	if !(a.Location.IsRoom && a.Location.TargetType == a.Type) {
		return false
	}
	
	// In the correct room, but only done if all room spaces below this
	// one are also occupied by an amphipod of the same type.
	at := a.Location.Down
	for at != nil {
		if at.Occupant == nil || at.Occupant.Type != a.Type {
			return false
		}
		at = at.Down
	}

	return true
}

type Move struct {
	Amphipod *Amphipod
	From, To *Space
	Cost int
}

func (m *Move) Do() {
	m.From.Occupant = nil
	m.To.Occupant = m.Amphipod
	m.Amphipod.Location = m.To
}

func (m *Move) Undo() {
	m.To.Occupant = nil
	m.From.Occupant = m.Amphipod
	m.Amphipod.Location = m.From
}

func (m *Move) String() string {
	return fmt.Sprintf("Move %s from %s to %s for %d energy", m.Amphipod, m.From.ID(), m.To.ID(), m.Cost)
}

type MoveSequence = Sequence[*Move]

func TotalCost(seq *MoveSequence) int {
	return streams.SumFunc(seq.Stream(), func(m *Move) int { return m.Cost })
}

// Options returns a slice of the locations which this amphipod is legally
// able to move to given the rules.
func (a *Amphipod) Options() []*Move {
	if a.IsDone() {
		return nil
	}

	unitCost := MoveCosts[a.Type]

	if a.Location.IsRoom {
		// In a starting position. Go up to an unoccupied space in the hallway.
		aboveRoom := a.Location.Up
		aboveRoomCost := unitCost
		for aboveRoom.IsRoom {
			if aboveRoom.Occupant != nil {
				// In a lower room space with another amphipod blocking the way.
				return nil
			}
			aboveRoomCost += unitCost
			aboveRoom = aboveRoom.Up
		}

		options := make([]*Move, 0, 7)

		// Now in the hallway space above the room which cannot be occupied.
		// Move left and right to look for options to move to.
		to := aboveRoom.Left
		cost := aboveRoomCost + unitCost
		// While we don't reach the end or bump into another amphipod...
		for to != nil && to.Occupant == nil {
			if to.Down == nil { // Not also above a room.
				options = append(options, &Move{Amphipod: a, From: a.Location, To: to, Cost: cost})
			}
			// Step left again.
			to = to.Left
			cost += unitCost
		}

		to = aboveRoom.Right
		cost = aboveRoomCost + unitCost
		// While we don't reach the end or bump into another amphipod...
		for to != nil && to.Occupant == nil {
			if to.Down == nil { // Not also above a room.
				options = append(options, &Move{Amphipod: a, From: a.Location, To: to, Cost: cost})
			}
			// Step right again.
			to = to.Right
			cost += unitCost
		}

		return options
	}

	// Currently in the hallway. Our only option is to move to a target room space if it is not blocked.
	var targetRoom *Space

	// Search left.
	to := a.Location.Left
	cost := unitCost
	// While we don't reach the end or bump into another amphipod...
	for to != nil && to.Occupant == nil {
		if to.Down != nil && to.Down.IsRoom && to.Down.TargetType == a.Type {
			// Found target room.
			targetRoom = to.Down
			break
		}

		// Step left again.
		to = to.Left
		cost += unitCost
	}

	if targetRoom == nil {
		// Search Right.
		to = a.Location.Right
		cost = unitCost
		// While we don't reach the end or bump into another amphipod...
		for to != nil && to.Occupant == nil {
			if to.Down != nil && to.Down.IsRoom && to.Down.TargetType == a.Type {
				// Found target room
				targetRoom = to.Down
				break
			}

			// Step right again.
			to = to.Right
			cost += unitCost
		}
	}

	if targetRoom == nil || targetRoom.Occupant != nil {
		return nil // Blocked with no options.
	}

	to = targetRoom
	cost += unitCost

	// Go down into the room as far as possible.
	for to.Down != nil && to.Down.Occupant == nil {
		to = to.Down
		cost += unitCost
	}

	if to.Down != nil && !to.Down.Occupant.IsDone() {
		return nil // Blocked by an occupant that needs to get out first.
	}

	return []*Move{&Move{Amphipod: a, From: a.Location, To: to, Cost: cost}}
}

type Configuration struct {
	Hallway *Space
	Rooms []*Space
	Amphipods []*Amphipod
}

func NewConfiguration(initialTypes []AmphipodType) *Configuration {
	// Build the hallway and rooms and link them all together.
	hallway := &Space{}
	for i := 1; i < 11; i++ { // 11 hallway spaces.
		// Build the hallway from right to left. The first and
		// far-right space is already created above.
		hallway = &Space{
			Right: hallway,
		}
		hallway.Right.Left = hallway
	}

	// Create the upper room spaces.
	roomA := &Space{IsRoom: true, TargetType: Amber, Up: hallway.Right.Right}
	roomA.Up.Down = roomA
	roomB := &Space{IsRoom: true, TargetType: Bronze, Up: roomA.Up.Right.Right}
	roomB.Up.Down = roomB
	roomC := &Space{IsRoom: true, TargetType: Copper, Up: roomB.Up.Right.Right}
	roomC.Up.Down = roomC
	roomD := &Space{IsRoom: true, TargetType: Desert, Up: roomC.Up.Right.Right}
	roomD.Up.Down = roomD

	rooms := []*Space{roomA, roomB, roomC, roomD}

	var amphipods []*Amphipod
	row, initialTypes := initialTypes[:4], initialTypes[4:]
	for i, roomSpace := range rooms {
		amphipod := &Amphipod{Type: row[i], Location: roomSpace}
		roomSpace.Occupant = amphipod
		amphipods = append(amphipods, amphipod)
	}

	for len(initialTypes) > 0 {
		// Create another row of room spaces
		roomA = &Space{IsRoom: true, TargetType: Amber, Up: roomA}
		roomA.Up.Down = roomA
		roomB = &Space{IsRoom: true, TargetType: Bronze, Up: roomB}
		roomB.Up.Down = roomB
		roomC = &Space{IsRoom: true, TargetType: Copper, Up: roomC}
		roomC.Up.Down = roomC
		roomD = &Space{IsRoom: true, TargetType: Desert, Up: roomD}
		roomD.Up.Down = roomD

		newRooms := []*Space{roomA, roomB, roomC, roomD}

		rooms = append(rooms, newRooms...)
		row, initialTypes = initialTypes[:4], initialTypes[4:]
		for i, roomSpace := range newRooms {
			amphipod := &Amphipod{Type: row[i], Location: roomSpace}
			roomSpace.Occupant = amphipod
			amphipods = append(amphipods, amphipod)
		}
	}

	return &Configuration{hallway, rooms, amphipods}
}

func (c *Configuration) IsOrganized() bool {
	for _, amphipod := range c.Amphipods {
		if !amphipod.IsDone() {
			return false
		}
	}
	return true
}

func (c *Configuration) Fingerprint() string {
	var b strings.Builder
	at := c.Hallway
	for at != nil {
		b.WriteString(at.String())
		at = at.Right
	}
	for _, room := range c.Rooms {
		b.WriteString(room.String())
	}
	return b.String()
}

type ConfigCostCache map[string]*MoveSequence

func (c *Configuration) LeastCostMovesToOrganize(cache ConfigCostCache) *MoveSequence {
	fingerprint := c.Fingerprint()

	if cachedMoves, ok := cache[fingerprint]; ok {
		return cachedMoves
	}

	if c.IsOrganized() {
		seq := &MoveSequence{}
		cache[fingerprint] = seq
		return seq
	}

	allOptions := streams.FlatMap(streams.FromItems(c.Amphipods...), func(amphipod *Amphipod) streams.Stream[*Move] {
		return streams.FromItems(amphipod.Options()...)
	})

	evaluatedOptions := streams.Map(allOptions, func(option *Move) *MoveSequence {
		option.Do()
		defer option.Undo()
		seq := c.LeastCostMovesToOrganize(cache)
		if seq != nil {
			seq = seq.ShallowCopy()
			seq.Prepend(option)
		}
		return seq
	})

	evaluatedOptions = streams.Filter(evaluatedOptions, func(moves *MoveSequence) bool {
		return moves != nil
	})

	minCostMoves := streams.MinFunc(evaluatedOptions, func(a, b *MoveSequence) bool { return TotalCost(a) < TotalCost(b) })

	cache[fingerprint] = minCostMoves
	return minCostMoves
}

func main() {
	initialTypes := loadAmphipodTypes()
	fmt.Println(initialTypes)
	config := NewConfiguration(initialTypes)

	leastCostMoves := config.LeastCostMovesToOrganize(ConfigCostCache{})
	fmt.Printf("Part 1: %d\n", TotalCost(leastCostMoves))
	streams.ForEach(leastCostMoves.Stream(), func(move *Move) { fmt.Println(move) })

	insertTypes := []AmphipodType{Desert, Copper, Bronze, Amber, Desert, Bronze, Amber, Copper}
	initialTypes = append(initialTypes[:4], append(insertTypes, initialTypes[4:]...)...)
	fmt.Println(initialTypes)
	config = NewConfiguration(initialTypes)

	leastCostMoves = config.LeastCostMovesToOrganize(ConfigCostCache{})
	fmt.Printf("Part 2: %d\n", TotalCost(leastCostMoves))
	streams.ForEach(leastCostMoves.Stream(), func(move *Move) { fmt.Println(move) })
}
