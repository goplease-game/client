package ds

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type HexCoord struct {
	Q int `json:"q"`
	R int `json:"r"`
}

type BoardCell struct {
	Coord      HexCoord `json:"coord"`
	Unit       *Unit    `json:"unit"`
	IsSafeZone bool     `json:"is_safe_zone"`
}

type BoardCells map[HexCoord]*BoardCell

type Board struct {
	Cells BoardCells `json:"cells"`
}

func (b *BoardCells) UnmarshalJSON(data []byte) error {
	tmp := make(map[string]*BoardCell)
	if err := json.Unmarshal(data, &tmp); err != nil {
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

// Needed for saving state
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
