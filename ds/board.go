package ds

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// HexCoord represents a position on the hex board using axial coordinates.
type HexCoord struct {
	Q int `json:"q"`
	R int `json:"r"`
}

// String returns the coordinate as a "(q,r)" string.
func (c HexCoord) String() string {
	return fmt.Sprintf("(%d,%d)", c.Q, c.R)
}

// BoardCell represents a single cell on the board, optionally occupied by a unit.
type BoardCell struct {
	Coord      HexCoord `json:"coord"`
	Unit       *Unit    `json:"unit"`
	IsSafeZone bool     `json:"is_safe_zone"`
}

// BoardCells maps hex coordinates to the cells on the board.
type BoardCells map[HexCoord]*BoardCell

// Board represents the hex board and its cells.
type Board struct {
	Cells BoardCells `json:"cells"`
}

// UnmarshalJSON decodes BoardCells from a JSON object keyed by "q:r" coordinate strings.
func (b *BoardCells) UnmarshalJSON(data []byte) error {
	tmp := make(map[string]*BoardCell)
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	if *b == nil {
		*b = make(BoardCells, len(tmp))
	}

	for k, cell := range tmp {
		parts := strings.Split(k, ":")
		q, _ := strconv.Atoi(parts[0])
		r, _ := strconv.Atoi(parts[1])
		coord := HexCoord{Q: q, R: r}

		(*b)[coord] = cell
	}

	return nil
}

// MarshalJSON encodes BoardCells as a JSON object keyed by "q:r" coordinate strings.
func (b BoardCells) MarshalJSON() ([]byte, error) {
	type Alias BoardCell

	out := make(map[string]*BoardCell, len(b))
	for coord, cell := range b {
		if cell == nil {
			continue
		}

		key := fmt.Sprintf("%d:%d", coord.Q, coord.R)

		out[key] = cell
	}

	return json.Marshal(out)
}

// InBounds checks if the given hex coordinate exists within the board boundaries.
func (b *Board) InBounds(coord HexCoord) bool {
	_, exists := b.Cells[coord]
	return exists
}
