package arena

import "github.com/ebitenui/ebitenui/widget"

type ChildAdder interface {
	AddChild(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc
	AddToUnitLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc
	AddToHUDLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc
	AddToFXLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc
}

// ContainerChildAdder wraps *widget.Container to implement ChildAdder.
// All children go to the default layer since containers have no z-index concept.
type ContainerChildAdder struct {
	c *widget.Container
}

func NewContainerChildAdder(c *widget.Container) *ContainerChildAdder {
	return &ContainerChildAdder{c: c}
}

func (a *ContainerChildAdder) AddChild(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return a.c.AddChild(children...)
}

func (a *ContainerChildAdder) AddToUnitLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return a.c.AddChild(children...)
}

func (a *ContainerChildAdder) AddToHUDLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return a.c.AddChild(children...)
}

func (a *ContainerChildAdder) AddToFXLayer(children ...widget.PreferredSizeLocateableWidget) widget.RemoveChildFunc {
	return a.c.AddChild(children...)
}
