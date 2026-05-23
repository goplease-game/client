package ds

import (
	"encoding/json"
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
