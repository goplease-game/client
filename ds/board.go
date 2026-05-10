package ds

type Board [][]*BoardCell // [row][col]cell

type BoardCell struct {
	Unit       *Unit `json:"unit"`
	IsSafeZone bool  `json:"is_safe_zone"`
}
