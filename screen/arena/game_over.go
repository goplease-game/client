package arena

import (
	"fmt"
	"image/color"
	"strconv"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/goplease-game/client/ui"
	server "github.com/goplease-game/server"
	sds "github.com/goplease-game/server/ds"
	"golang.org/x/image/colornames"
)

// Colors for the You/Opponent comparison bars.
var (
	gameOverYouBarTrack = color.NRGBA{60, 60, 60, 255}
	gameOverYouBarFill  = color.NRGBA{80, 180, 255, 255}
	gameOverOppBarTrack = color.NRGBA{60, 60, 60, 255}
	gameOverOppBarFill  = color.NRGBA{255, 120, 90, 255}

	gameOverStatBarWidth   = 140
	gameOverStatBarHeight  = 12
	gameOverStatLabelWidth = 150
	gameOverStatValueWidth = 36
)

// gameOverStatRows defines the ordered list of stats shown side-by-side on the game-over screen.
var gameOverStatRows = []struct {
	label string
	get   func(*server.PlayerStats) int
}{
	{"Damage dealt", func(ps *server.PlayerStats) int { return ps.DamageDealt }},
	{"Damage received", func(ps *server.PlayerStats) int { return ps.DamageReceived }},
	{"Overkill damage", func(ps *server.PlayerStats) int { return ps.OverkillDamage }},
	{"Kills", func(ps *server.PlayerStats) int { return ps.Kills }},
	{"Shield applied", func(ps *server.PlayerStats) int { return ps.ShieldApplied }},
	{"Shield destroyed", func(ps *server.PlayerStats) int { return ps.ShieldDestroyed }},
	{"Healing received", func(ps *server.PlayerStats) int { return ps.HealingReceived }},
	{"Healing wasted", func(ps *server.PlayerStats) int { return ps.HealingWasted }},
	{"Cells traveled", func(ps *server.PlayerStats) int { return ps.CellsTraveled }},
	{"Abilities used", func(ps *server.PlayerStats) int { return ps.AbilitiesUsed }},
}

// showGameOverOverlay displays the game-over overlay with the given title and match statistics.
// Creates the overlay lazily on first call.
func (s *Screen) showGameOverOverlay(win bool, explain string, stats *server.Stats, localPlayerID string) {
	title := "You Lose"
	titleColor := gameOverLoseColor
	if win {
		title = "You Win"
		titleColor = gameOverWinColor
	}

	overlay := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{0, 0, 0, 180})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			}),
		),
	)

	panel := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(20),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(40)),
		)),
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(headerBgColor)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	tf := ui.TextFaceBold(48)
	panel.AddChild(widget.NewText(
		widget.TextOpts.Text(title, &tf, titleColor),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	))

	if explain != "" {
		tf := ui.TextFaceBold(20)
		panel.AddChild(widget.NewText(
			widget.TextOpts.Text(explain, &tf, colornames.Whitesmoke),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		))
	}

	if stats != nil {
		panel.AddChild(s.gameOverMetaPanel(stats, localPlayerID))
		panel.AddChild(s.gameOverStatsGrid(stats, localPlayerID))
	}

	buttons := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(12),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	buttons.AddChild(s.menuButton("Play Again", func(_ *widget.ButtonClickedEventArgs) {
		if s.OnRestartScreen != nil {
			s.nextScreen = s.OnRestartScreen()
		} else {
			printD("Play Again: OnRestartScreen is not set")
		}
	}))
	buttons.AddChild(s.menuButton("Main Menu", func(_ *widget.ButtonClickedEventArgs) {
		s.nextScreen = s.OnExitScreen()
	}))

	panel.AddChild(buttons)
	overlay.AddChild(panel)

	s.gameOverUI = &ebitenui.UI{Container: overlay}
	s.gameOverVisible = true
}

// gameOverMetaPanel builds the round/duration/first-blood summary line above the stats table.
func (s *Screen) gameOverMetaPanel(stats *server.Stats, localPlayerID string) *widget.Container {
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(4),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	tf := ui.TextFaceBold(16)

	c.AddChild(widget.NewText(
		widget.TextOpts.Text(
			fmt.Sprintf("Round %d  •  %s", stats.RoundNumber, formatDuration(stats.Duration())),
			&tf, colornames.Whitesmoke,
		),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Position: widget.RowLayoutPositionCenter}),
		),
	))

	if fb := firstBloodText(stats.FirstBlood, localPlayerID); fb != "" {
		c.AddChild(widget.NewText(
			widget.TextOpts.Text(fb, &tf, colornames.Whitesmoke),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{Position: widget.RowLayoutPositionCenter}),
			),
		))
	}

	return c
}

// gameOverStatsGrid builds the mirrored You/Opponent stats comparison, one row per stat,
// followed by a final Score row. Each row has two bars growing outward from a center label:
// the larger value's bar fills its track completely, the smaller value's bar is sized
// relative to it.
func (s *Screen) gameOverStatsGrid(stats *server.Stats, localPlayerID string) *widget.Container {
	you, opp := splitPlayerStats(stats, localPlayerID)

	grid := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(8),
		)),
	)

	for _, row := range gameOverStatRows {
		youVal := row.get(you)
		oppVal := row.get(opp)
		grid.AddChild(s.gameOverStatRow(row.label, youVal, oppVal))
	}

	grid.AddChild(s.gameOverScoreRow(you.Score, opp.Score))

	return grid
}

// gameOverScoreRow builds the final Score summary row: large centered numbers on each side
// of the "Score" label, without comparison bars since this is a derived total rather than a
// raw stat.
func (s *Screen) gameOverScoreRow(youScore, oppScore int) *widget.Container {
	row := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(10),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)

	scoreFace := ui.TextFaceBold(22)
	labelFace := ui.TextFaceBold(14)

	row.AddChild(widget.NewText(
		widget.TextOpts.Text(strconv.Itoa(youScore), &scoreFace, gameOverYouBarFill),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(gameOverStatBarWidth+gameOverStatValueWidth, 0),
		),
	))

	row.AddChild(widget.NewText(
		widget.TextOpts.Text("SCORE", &labelFace, colornames.Whitesmoke),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(gameOverStatLabelWidth, 0),
		),
	))

	row.AddChild(widget.NewText(
		widget.TextOpts.Text(strconv.Itoa(oppScore), &scoreFace, gameOverOppBarFill),
		widget.TextOpts.Text(strconv.Itoa(oppScore), &scoreFace, gameOverOppBarFill),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(gameOverStatBarWidth+gameOverStatValueWidth, 0),
		),
	))

	return row
}

// gameOverStatRow builds a single row: [you bar] (you value) (label) (opp value) [opp bar].
// Both bars grow outward from the shared center toward their value's edge, so longer bars are
// immediately readable as "more" for that stat, while both numbers and the label stay together
// in the middle so the eye doesn't have to travel to the screen edges.
func (s *Screen) gameOverStatRow(label string, youVal, oppVal int) *widget.Container {
	row := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(10),
		)),
	)

	valFace := ui.TextFaceBold(14)
	labelFace := ui.TextFaceBold(14)

	row.AddChild(s.gameOverStatBar(youVal, oppVal, gameOverYouBarTrack, gameOverYouBarFill, true))

	row.AddChild(widget.NewText(
		widget.TextOpts.Text(strconv.Itoa(youVal), &valFace, gameOverYouBarFill),
		widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(gameOverStatValueWidth, 0),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Position: widget.RowLayoutPositionCenter}),
		),
	))

	row.AddChild(widget.NewText(
		widget.TextOpts.Text(label, &labelFace, colornames.Whitesmoke),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(gameOverStatLabelWidth, 0),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Position: widget.RowLayoutPositionCenter}),
		),
	))

	row.AddChild(widget.NewText(
		widget.TextOpts.Text(strconv.Itoa(oppVal), &valFace, gameOverOppBarFill),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(gameOverStatValueWidth, 0),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Position: widget.RowLayoutPositionCenter}),
		),
	))

	row.AddChild(s.gameOverStatBar(oppVal, youVal, gameOverOppBarTrack, gameOverOppBarFill, false))

	return row
}

// gameOverStatBar builds a single horizontal progress bar sized relative to the larger of
// value and other. When inverted is true, the fill grows from the bar's right edge leftward
// (used for the "You" side, so it grows toward the center label); otherwise it grows from the
// left edge rightward (used for the "Opponent" side, growing away from the center label).
func (s *Screen) gameOverStatBar(value, other int, trackColor, fillColor color.NRGBA, inverted bool) *widget.ProgressBar {
	maxVal := max(value, other)
	maxVal = max(maxVal, 1)

	return widget.NewProgressBar(
		widget.ProgressBarOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(gameOverStatBarWidth, gameOverStatBarHeight),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Position: widget.RowLayoutPositionCenter}),
		),
		widget.ProgressBarOpts.Inverted(inverted),
		widget.ProgressBarOpts.Images(
			&widget.ProgressBarImage{
				Idle: image.NewNineSliceColor(trackColor),
			},
			&widget.ProgressBarImage{
				Idle: image.NewNineSliceColor(fillColor),
			},
		),
		widget.ProgressBarOpts.Values(0, maxVal, value),
		widget.ProgressBarOpts.TrackPadding(&widget.Insets{
			Top: 1, Bottom: 1, Left: 1, Right: 1,
		}),
	)
}

// splitPlayerStats returns the local player's and the opponent's PlayerStats,
// falling back to empty stats if either side is missing from the map.
func splitPlayerStats(stats *server.Stats, localPlayerID string) (you, opp *server.PlayerStats) {
	pID, _ := sds.ParseID(localPlayerID)
	you = stats.Players[pID]
	for id, ps := range stats.Players {
		if id != pID {
			opp = ps
			break
		}
	}

	if you == nil {
		you = &server.PlayerStats{}
	}
	if opp == nil {
		opp = &server.PlayerStats{}
	}

	return you, opp
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := d / time.Minute
	sec := (d % time.Minute) / time.Second

	return fmt.Sprintf("%02d:%02d", m, sec)
}

// firstBloodText renders a human-readable first-blood summary relative to localPlayerID.
func firstBloodText(fb *server.FirstBloodStats, localPlayerID string) string {
	if fb == nil {
		return ""
	}

	killerPossessive := "Opponent's"
	if fb.KillerSide.String() == localPlayerID {
		killerPossessive = "Your"
	}
	victimPossessive := "opponent's"
	if fb.VictimSide.String() == localPlayerID {
		victimPossessive = "your"
	}

	return fmt.Sprintf("First blood: %s %s defeated %s %s (Round %d)",
		killerPossessive, fb.KillerName, victimPossessive, fb.VictimName, fb.Round)
}
