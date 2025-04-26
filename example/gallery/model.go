// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import "github.com/hajimehoshi/guigui/basicwidget"

type Model struct {
	mode string

	texts      TextsModel
	textFields TextFieldsModel
}

func (m *Model) Mode() string {
	if m.mode == "" {
		return "settings"
	}
	return m.mode
}

func (m *Model) SetMode(mode string) {
	m.mode = mode
}

func (m *Model) Texts() *TextsModel {
	return &m.texts
}

func (m *Model) TextFields() *TextFieldsModel {
	return &m.textFields
}

type TextsModel struct {
	text    string
	textSet bool

	horizontalAlign basicwidget.HorizontalAlign
	verticalAlign   basicwidget.VerticalAlign
	noWrap          bool
	bold            bool
	selectable      bool
	editable        bool
}

func (t *TextsModel) Text() string {
	if !t.textSet {
		return `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
隴西の李徴は博学才穎、天宝の末年、若くして名を虎榜に連ね、ついで江南尉に補せられたが、性、狷介、自ら恃むところ頗る厚く、賤吏に甘んずるを潔しとしなかった。`
	}
	return t.text
}

func (t *TextsModel) SetText(text string) {
	t.text = text
	t.textSet = true
}

func (t *TextsModel) HorizontalAlign() basicwidget.HorizontalAlign {
	return t.horizontalAlign
}

func (t *TextsModel) SetHorizontalAlign(align basicwidget.HorizontalAlign) {
	t.horizontalAlign = align
}

func (t *TextsModel) VerticalAlign() basicwidget.VerticalAlign {
	return t.verticalAlign
}

func (t *TextsModel) SetVerticalAlign(align basicwidget.VerticalAlign) {
	t.verticalAlign = align
}

func (t *TextsModel) AutoWrap() bool {
	return !t.noWrap
}

func (t *TextsModel) SetAutoWrap(autoWrap bool) {
	t.noWrap = !autoWrap
}

func (t *TextsModel) Bold() bool {
	return t.bold
}

func (t *TextsModel) SetBold(bold bool) {
	t.bold = bold
}

func (t *TextsModel) Selectable() bool {
	return t.selectable
}

func (t *TextsModel) SetSelectable(selectable bool) {
	t.selectable = selectable
	if !selectable {
		t.editable = false
	}
}

func (t *TextsModel) Editable() bool {
	return t.editable
}

func (t *TextsModel) SetEditable(editable bool) {
	t.editable = editable
	if editable {
		t.selectable = true
	}
}

type TextFieldsModel struct {
	singleLineText     string
	singleLinetTextSet bool
	multilineText      string
	multilineTextSet   bool

	horizontalAlign basicwidget.HorizontalAlign
	verticalAlign   basicwidget.VerticalAlign
	autoWrap        bool
	disabled        bool
}

func (t *TextFieldsModel) SingleLineText() string {
	if !t.singleLinetTextSet {
		return "Hello, Guigui!"
	}
	return t.singleLineText
}

func (t *TextFieldsModel) SetSingleLineText(text string) {
	t.singleLineText = text
	t.singleLinetTextSet = true
}

func (t *TextFieldsModel) MultilineText() string {
	if !t.multilineTextSet {
		return "Hello, Guigui!\nThis is a multiline text field."
	}
	return t.multilineText
}

func (t *TextFieldsModel) SetMultilineText(text string) {
	t.multilineText = text
	t.multilineTextSet = true
}

func (t *TextFieldsModel) HorizontalAlign() basicwidget.HorizontalAlign {
	return t.horizontalAlign
}

func (t *TextFieldsModel) SetHorizontalAlign(align basicwidget.HorizontalAlign) {
	t.horizontalAlign = align
}

func (t *TextFieldsModel) VerticalAlign() basicwidget.VerticalAlign {
	return t.verticalAlign
}

func (t *TextFieldsModel) SetVerticalAlign(align basicwidget.VerticalAlign) {
	t.verticalAlign = align
}

func (t *TextFieldsModel) AutoWrap() bool {
	return t.autoWrap
}

func (t *TextFieldsModel) SetAutoWrap(autoWrap bool) {
	t.autoWrap = autoWrap
}

func (t *TextFieldsModel) Enabled() bool {
	return !t.disabled
}

func (t *TextFieldsModel) SetEnabled(enabled bool) {
	t.disabled = !enabled
}
