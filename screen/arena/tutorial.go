package arena

import (
	"fmt"
	"image/color"
	"log"
	"strings"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/goplease-game/client/config"
	"github.com/goplease-game/client/tutorial"
)

func (s *Screen) setupTutorial() {
	if len(s.snapshot.Tutorial.Steps) == 0 {
		return
	}

	conf := config.Get()
	if conf.SkipTutorial || tutorialCompleted(conf, s.snapshot.Tutorial.Name) {
		return
	}

	s.tutorialOverlay = NewTutorialOverlay(s.snapshot.Tutorial,
		func() {
			s.clearTutorialHighlights()
			config.Get().SkipTutorial = true
			if err := config.Save(); err != nil {
				log.Printf("tutorial: failed to save skip state: %v", err)
			}
			s.tutorialOverlay = nil
			s.refreshStatusBar()
		},
		func() {
			s.clearTutorialHighlights()
			markTutorialCompleted(s.snapshot.Tutorial.Name)
			s.tutorialOverlay = nil
			s.refreshStatusBar()
		},
		s.applyTutorialStep,
		func(visible bool) {
			if !visible {
				s.clearTutorialHighlights()
			}
			s.refreshStatusBar()
		},
	)
	s.refreshStatusBar()
}

func (s *Screen) applyTutorialStep(step tutorial.Step) {
	s.clearTutorialHighlights()

	switch step.Highlight {
	case tutorial.HighlightUnitPanel:
		if s.unitPanelRef != nil {
			s.unitPanelRef.SetBackgroundImage(tutorialHighlightImage(unitPanelBgColor))
		}
	case tutorial.HighlightQueue:
		if s.queuePanelRef != nil {
			s.queuePanelRef.SetBackgroundImage(tutorialHighlightImage(unitPanelBgColor))
		}
	case tutorial.HighlightAbilityPanel:
		if s.abilityPanelRef != nil {
			s.abilityPanelRef.SetBackgroundImage(tutorialHighlightImage(footerBgColor))
		}
	case tutorial.HighlightEndTurn:
		if img := s.nextActionButtonImage(); img != nil {
			img.Idle = tutorialHighlightImage(color.NRGBA{0x22, 0x8B, 0x22, 0xff})
			img.Hover = tutorialHighlightImage(color.NRGBA{0x32, 0xAB, 0x32, 0xff})
		}
	}
}

func (s *Screen) refreshTutorialStep() {
	if s.tutorialOverlay != nil && s.tutorialOverlay.IsVisible() {
		step, ok := s.tutorialOverlay.CurrentStep()
		if !ok {
			return
		}
		s.applyTutorialStep(step)
	}
}

func (s *Screen) clearTutorialHighlights() {
	if s.unitPanelRef != nil {
		s.unitPanelRef.SetBackgroundImage(image.NewNineSliceColor(unitPanelBgColor))
	}
	if s.queuePanelRef != nil {
		s.queuePanelRef.SetBackgroundImage(image.NewNineSliceColor(unitPanelBgColor))
	}
	if s.abilityPanelRef != nil {
		s.abilityPanelRef.SetBackgroundImage(image.NewNineSliceColor(footerBgColor))
	}
	if img := s.nextActionButtonImage(); img != nil && !s.endTurnBtnPulseActive {
		img.Idle = endTurnBtnIdle()
		img.Hover = endTurnBtnHover()
	}
}

func (s *Screen) nextActionButtonImage() *widget.ButtonImage {
	if s.nextActionBtn == nil {
		return nil
	}
	return s.nextActionBtn.Image()
}

func tutorialCompleted(conf *config.Config, name string) bool {
	for _, completed := range conf.TutorialsCompleted {
		if completed == name {
			return true
		}
	}
	return false
}

func markTutorialCompleted(name string) {
	conf := config.Get()
	if tutorialCompleted(conf, name) {
		return
	}

	conf.TutorialsCompleted = append(conf.TutorialsCompleted, name)
	if err := config.Save(); err != nil {
		log.Printf("tutorial: failed to save completion state: %v", err)
	}
}

func tutorialHighlightImage(fill color.Color) *image.NineSlice {
	return image.NewBorderedNineSliceColor(
		fill,
		color.NRGBA{0xff, 0xd2, 0x4a, 0xff},
		3,
	)
}

// parseTutorialMessage splits a message string into text and image segments.
// Image tags have the form [@pic:filename.png;WxH].
func parseTutorialMessage(msg string) []tutorialMessageSegment {
	var segments []tutorialMessageSegment
	rest := msg
	for {
		start := strings.Index(rest, "[@pic:")
		if start == -1 {
			if rest != "" {
				segments = append(segments, tutorialMessageSegment{text: rest})
			}
			break
		}
		if start > 0 {
			segments = append(segments, tutorialMessageSegment{text: rest[:start]})
		}
		end := strings.Index(rest[start:], "]")
		if end == -1 {
			segments = append(segments, tutorialMessageSegment{text: rest})
			break
		}
		tag := rest[start+6 : start+end] // content after [@pic:

		var filename string
		var w, h int
		if idx := strings.Index(tag, ";"); idx != -1 {
			filename = tag[:idx]
			fmt.Sscanf(tag[idx+1:], "%dx%d", &w, &h)
		} else {
			filename = tag
		}

		segments = append(segments, tutorialMessageSegment{image: filename, imgW: w, imgH: h})
		rest = rest[start+end+1:]
	}
	return segments
}

type tutorialMessageSegment struct {
	text  string
	image string
	imgW  int
	imgH  int
}
