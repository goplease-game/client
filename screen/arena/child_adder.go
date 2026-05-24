package arena

import "github.com/ebitenui/ebitenui/widget"

// ChildAdder is implemented by any widget that supports layered child widgets.
// It abstracts over HexCellWidget (which has z-index layers) and
// ContainerChildAdder (which wraps a flat *widget.Container for UI panels).
type ChildAdder interface {
	AddChild(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc
	AddToUnitLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc
	AddToHUDLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc
	AddToFXLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc
}

// ContainerChildAdder wraps *widget.Container to implement ChildAdder.
// Used for UI panels (queue cards, ability cards) that have no z-index concept —
// all layer methods delegate to the same underlying AddChild.
type ContainerChildAdder struct {
	c *widget.Container
}

// NewContainerChildAdder wraps c so it can be passed to buildBoardCard and similar
// functions that accept a ChildAdder.
func NewContainerChildAdder(c *widget.Container) *ContainerChildAdder {
	return &ContainerChildAdder{c: c}
}

// AddChild adds widgets directly to the container.
func (a *ContainerChildAdder) AddChild(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return a.c.AddChild(children...)
}

// AddToUnitLayer delegates to AddChild — containers have no layer concept.
func (a *ContainerChildAdder) AddToUnitLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return a.c.AddChild(children...)
}

// AddToHUDLayer delegates to AddChild — containers have no layer concept.
func (a *ContainerChildAdder) AddToHUDLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return a.c.AddChild(children...)
}

// AddToFXLayer delegates to AddChild — containers have no layer concept.
func (a *ContainerChildAdder) AddToFXLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return a.c.AddChild(children...)
}
