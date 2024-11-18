package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SelectableCard struct {
	widget.BaseWidget
	card     *widget.Card
	border   *canvas.Rectangle
	selected bool
	onSelect func(selected bool)
}

func NewSelectableCard(title, subtitle string, image fyne.Resource, buttonLabel string, buttonFunc func(), onSelect func(selected bool)) *SelectableCard {
	img := canvas.NewImageFromResource(image)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(100, 100)) // Set a smaller size for the image

	button := widget.NewButton(buttonLabel, buttonFunc) // Use the button function parameter here

	cardContent := container.NewVBox(
		img,
		widget.NewLabel(title),
		widget.NewLabel(subtitle),
		button, // Add the button here
	)

	card := widget.NewCard("", "", cardContent)
	border := canvas.NewRectangle(color.Transparent)
	border.StrokeColor = color.Transparent
	border.StrokeWidth = 0

	selectableCard := &SelectableCard{
		card:     card,
		border:   border,
		selected: false,
		onSelect: onSelect,
	}

	selectableCard.ExtendBaseWidget(selectableCard)
	return selectableCard
}

func (c *SelectableCard) Tapped(_ *fyne.PointEvent) {
	c.selected = !c.selected
	if c.selected {
		c.border.StrokeColor = color.RGBA{255, 255, 0, 255} // Yellow color
		c.border.StrokeWidth = 4
	} else {
		c.border.StrokeColor = color.Transparent
		c.border.StrokeWidth = 0
	}
	c.border.Refresh()
	c.Refresh()
	c.onSelect(c.selected) // Call the onSelect function with the selected state
}

func (c *SelectableCard) CreateRenderer() fyne.WidgetRenderer {
	objects := []fyne.CanvasObject{c.border, c.card}
	return &cardRenderer{
		card:    c,
		objects: objects,
	}
}

type cardRenderer struct {
	card    *SelectableCard
	objects []fyne.CanvasObject
}

func (r *cardRenderer) Layout(size fyne.Size) {
	padding := fyne.NewSize(8, 8)
	r.objects[1].Resize(size.Subtract(padding))
	r.objects[1].Move(fyne.NewPos(padding.Width/2, padding.Height/2))
	r.objects[0].Resize(r.objects[1].Size().Add(padding))
	r.objects[0].Move(fyne.NewPos(0, 0))
}

func (r *cardRenderer) MinSize() fyne.Size {
	padding := fyne.NewSize(8, 8)
	return r.objects[1].MinSize().Add(padding)
}

func (r *cardRenderer) Refresh() {
	r.objects[0].Refresh()
	r.objects[1].Refresh()
}

func (r *cardRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (r *cardRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *cardRenderer) Destroy() {}
