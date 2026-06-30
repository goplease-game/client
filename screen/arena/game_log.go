package arena

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/goplease-game/client/config"
	"github.com/goplease-game/client/ui"
)

// logMessage is the client-side representation of a server GameLog message.
type logMessage struct {
	Time      time.Time `json:"time"`
	Kind      string    `json:"kind"`
	Text      string    `json:"text"`
	Sender    string    `json:"sender,omitempty"`
	Recipient string    `json:"recipient,omitempty"`
}

// logTagColors maps semantic tag names to BBCode hex colors.
var logTagColors = map[string]string{
	"ability": logAbilityTextColor,
	"damage":  logDamageColor,
	"shield":  logShieldColor,
	"round":   logRoundColor,
	"hp":      logHPColor,
	"ap":      logAPColor,
}

// unitTagRe matches <unit id="...">name</unit> produced by the server.
var unitTagRe = regexp.MustCompile(`<unit id="([^"]+)">([^<]+)</unit>`)
var playerTagRe = regexp.MustCompile(`<player id="([^"]+)">([^<]+)</player>`)

// toBBCode converts server semantic tags to EbitenUI BBCode, resolving
// friend/enemy coloring directly against the screen's current game state.
func (s *Screen) toBBCode(text string) string {
	text = unitTagRe.ReplaceAllStringFunc(text, func(match string) string {
		groups := unitTagRe.FindStringSubmatch(match)
		if len(groups) < 3 {
			return match
		}
		id, name := groups[1], groups[2]
		col := logEnemyColor
		if s.isMyUnit(id) {
			col = logFriendlyColor
		}
		return fmt.Sprintf("[color=%s]%s[/color]", col, name)
	})

	text = playerTagRe.ReplaceAllStringFunc(text, func(match string) string {
		groups := playerTagRe.FindStringSubmatch(match)
		if len(groups) < 3 {
			return match
		}
		id, name := groups[1], groups[2]
		col := logEnemyColor
		if id == s.player.ID {
			col = logFriendlyColor
		}
		return fmt.Sprintf("[color=%s]%s[/color]", col, name)
	})

	for tag, col := range logTagColors {
		text = strings.ReplaceAll(text, "<"+tag+">", "[color="+col+"]")
		text = strings.ReplaceAll(text, "</"+tag+">", "[/color]")
	}

	return text
}

// logKindPrefix returns a short BBCode prefix for a message kind.
func logKindPrefix(kind string) string {
	switch kind {
	case "action":
		return "[color=" + logActionPrefixColor + "]> [/color]"
	case "system":
		return "[color=" + logSystemPrefixColor + "]* [/color]"
	case "error":
		return "[color=" + logErrorPrefixColor + "]ERR: [/color]"
	case "chat":
		return "[color=" + logChatPrefixColor + "]~ [/color]"
	default:
		return ""
	}
}

// gameLogWindow holds the state of the floating game log window.
type gameLogWindow struct {
	textarea *widget.TextArea
	messages []logMessage
}

const logPanelW = 300

// createLogPanel creates the panel for game log.
func (s *Screen) createLogPanel() *widget.Container {
	panel := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(logPanelBgColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchVertical: true,
				Padding: &widget.Insets{
					Top:    headerH,
					Bottom: footerH + statusH,
				},
			}),
			widget.WidgetOpts.MinSize(logPanelW, config.Get().WindowH-footerH-headerH-statusH),
		),
	)

	textFace := ui.TextFace(14)
	textarea := widget.NewTextArea(
		widget.TextAreaOpts.ContainerOpts(
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					StretchHorizontal: true,
					StretchVertical:   true,
				}),
			),
		),
		widget.TextAreaOpts.ControlWidgetSpacing(2),
		widget.TextAreaOpts.ProcessBBCode(true),
		widget.TextAreaOpts.FontFace(&textFace),
		widget.TextAreaOpts.FontColor(logTextColor),
		widget.TextAreaOpts.TextPadding(*widget.NewInsetsSimple(5)),
		widget.TextAreaOpts.ShowVerticalScrollbar(),
		widget.TextAreaOpts.VerticalScrollMode(widget.ScrollEnd),
		widget.TextAreaOpts.ScrollContainerImage(&widget.ScrollContainerImage{
			Idle: image.NewNineSliceColor(logPanelBgColor),
			Mask: image.NewNineSliceColor(logPanelBgColor),
		}),
		widget.TextAreaOpts.SliderParams(&widget.SliderParams{
			MinHandleSize: new(3),
			TrackPadding:  widget.NewInsetsSimple(3),
			TrackImage: &widget.SliderTrackImage{
				Idle:  image.NewNineSliceColor(logScrollbarTrackColor),
				Hover: image.NewNineSliceColor(logScrollbarHoverColor),
			},
			HandleImage: &widget.ButtonImage{
				Idle:    image.NewNineSliceColor(logScrollbarIdleColor),
				Hover:   image.NewNineSliceColor(logScrollbarHandleHoverColor),
				Pressed: image.NewNineSliceColor(logScrollbarPressedColor),
			},
		}),
	)

	s.logWindow = &gameLogWindow{textarea: textarea}
	panel.AddChild(textarea)

	panel.GetWidget().SetVisibility(widget.Visibility_Hide)
	return panel
}

// toggleGameLog opens or closes the game log window.
func (s *Screen) toggleGameLog() {
	if s.logPanelRef.GetWidget().GetVisibility() == widget.Visibility_Show {
		s.logPanelRef.GetWidget().SetVisibility(widget.Visibility_Hide)
		s.boardContainerRef.GetWidget().LayoutData = widget.AnchorLayoutData{
			StretchHorizontal: true,
			StretchVertical:   true,
			Padding: &widget.Insets{
				Top:    headerH,
				Bottom: footerH + statusH,
			},
		}
	} else {
		s.logPanelRef.GetWidget().SetVisibility(widget.Visibility_Show)
		s.boardContainerRef.GetWidget().LayoutData = widget.AnchorLayoutData{
			StretchHorizontal: true,
			StretchVertical:   true,
			Padding: &widget.Insets{
				Top:    headerH,
				Bottom: footerH + statusH,
				Left:   logPanelW,
			},
		}
	}
}

// handleGameLog deserialises an incoming GameLogAction message and appends
// it to the log window.
func (s *Screen) handleGameLog(data json.RawMessage) {
	var msg logMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		log.Printf("[gamelog] unmarshal error: %v", err)
		return
	}
	s.logWindow.messages = append(s.logWindow.messages, msg)
	s.appendLogEntry(msg)
}

func (s *Screen) appendLogEntry(msg logMessage) {
	prefix := ""
	if !msg.Time.IsZero() {
		prefix = fmt.Sprintf("[color=%s]%s[/color] ", logTimestampColor, msg.Time.Format("15:04"))
	}

	line := prefix + logKindPrefix(msg.Kind) + s.toBBCode(msg.Text) + "\n"
	s.logWindow.textarea.AppendText(line)
}

// isMyUnit returns true if the given unit ID belongs to the local player.
func (s *Screen) isMyUnit(id string) bool {
	for _, u := range s.player.Units {
		if u.ID == id {
			return true
		}
	}
	for _, cell := range s.board.Cells {
		if cell != nil && cell.Unit != nil && cell.Unit.ID == id && !cell.Unit.IsOpponent {
			return true
		}
	}
	return false
}
