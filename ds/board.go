package ds

type Board [][]*BoardCell

type BoardCell struct {
	Unit       *Unit `json:"unit"`
	IsSafeZone bool  `json:"is_safe_zone"`
}
