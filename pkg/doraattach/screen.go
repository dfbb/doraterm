// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build !windows && !(linux && (mips || mips64))

package doraattach

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	xterm "github.com/gitpod-io/xterm-go"
)

type renderedCell struct {
	ch             string
	fg             uint32
	bg             uint32
	width          int
	underlineStyle xterm.UnderlineStyle
	underlineColor uint32
}

type Viewport struct {
	mu              sync.Mutex
	term            *xterm.Terminal
	offsetX         int
	offsetY         int
	width           int
	height          int
	cols            int
	rows            int
	lastCells       [][]renderedCell
	needsFullRedraw bool
	inAltScreen     bool
	lastCursorCode  int
	lastYBase       int
}

func newViewport(remoteRows, remoteCols, localWidth, localHeight int) *Viewport {
	term := xterm.New(
		xterm.WithCols(remoteCols),
		xterm.WithRows(remoteRows),
	)
	vp := &Viewport{
		term:            term,
		width:           localWidth,
		height:          localHeight,
		cols:            remoteCols,
		rows:            remoteRows,
		needsFullRedraw: true,
		lastCursorCode:  -1,
		lastYBase:       -1,
	}
	vp.offsetY = remoteRows - localHeight
	if vp.offsetY < 0 {
		vp.offsetY = 0
	}
	return vp
}

func (vp *Viewport) Write(data []byte) (int, error) {
	vp.mu.Lock()
	defer vp.mu.Unlock()
	for i := 0; i+3 < len(data); i++ {
		if data[i] == 0x1b && data[i+1] == '[' && data[i+2] == '2' && data[i+3] == 'J' {
			vp.needsFullRedraw = true
			break
		}
	}
	return vp.term.Write(data)
}

func (vp *Viewport) MoveUp(n int) {
	vp.mu.Lock()
	defer vp.mu.Unlock()
	vp.offsetY -= n
	vp.clampOffsets()
}

func (vp *Viewport) MoveDown(n int) {
	vp.mu.Lock()
	defer vp.mu.Unlock()
	vp.offsetY += n
	vp.clampOffsets()
}

func (vp *Viewport) MoveLeft(n int) {
	vp.mu.Lock()
	defer vp.mu.Unlock()
	vp.offsetX -= n
	vp.clampOffsets()
}

func (vp *Viewport) MoveRight(n int) {
	vp.mu.Lock()
	defer vp.mu.Unlock()
	vp.offsetX += n
	vp.clampOffsets()
}

func (vp *Viewport) Resize(newWidth, newHeight int) {
	vp.mu.Lock()
	defer vp.mu.Unlock()
	vp.width = newWidth
	vp.height = newHeight
	vp.clampOffsets()
	vp.needsFullRedraw = true
}

func (vp *Viewport) ForceFullRedraw() {
	vp.mu.Lock()
	defer vp.mu.Unlock()
	vp.needsFullRedraw = true
}

func (vp *Viewport) Reset() {
	vp.mu.Lock()
	defer vp.mu.Unlock()
	vp.term = xterm.New(
		xterm.WithCols(vp.cols),
		xterm.WithRows(vp.rows),
	)
	vp.inAltScreen = false
	vp.lastCursorCode = -1
	vp.lastYBase = -1
	vp.lastCells = nil
	vp.needsFullRedraw = true
}

func (vp *Viewport) InAltScreen() bool {
	vp.mu.Lock()
	defer vp.mu.Unlock()
	return vp.inAltScreen
}

func (vp *Viewport) clampOffsets() {
	maxX := vp.cols - vp.width
	if maxX < 0 {
		maxX = 0
	}
	maxY := vp.rows - vp.height
	if maxY < 0 {
		maxY = 0
	}
	minY := -vp.term.Buffer().YBase
	if vp.offsetX < 0 {
		vp.offsetX = 0
	} else if vp.offsetX > maxX {
		vp.offsetX = maxX
	}
	if vp.offsetY < minY {
		vp.offsetY = minY
	} else if vp.offsetY > maxY {
		vp.offsetY = maxY
	}
}

func (vp *Viewport) Render(w io.Writer) {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	ox, oy := vp.offsetX, vp.offsetY
	width, height := vp.width, vp.height
	if width <= 0 || height <= 0 {
		return
	}

	buf := vp.term.Buffer()
	yBase := buf.YBase
	cursorX := vp.term.CursorX()
	cursorY := vp.term.CursorY()

	fullRedraw := vp.needsFullRedraw
	if yBase != vp.lastYBase {
		if vp.offsetY < 0 {
			vp.offsetY -= yBase - vp.lastYBase
			vp.clampOffsets()
		}
		vp.lastYBase = yBase
		vp.lastCells = nil
		fullRedraw = true
	}
	if len(vp.lastCells) != height {
		vp.lastCells = make([][]renderedCell, height)
		fullRedraw = true
	}
	for i := range vp.lastCells {
		if len(vp.lastCells[i]) != width {
			vp.lastCells[i] = make([]renderedCell, width)
			fullRedraw = true
		}
	}
	vp.needsFullRedraw = false

	var out bytes.Buffer
	out.WriteString("\x1b[?25l")

	remoteAlt := vp.term.IsAltBufferActive()
	if remoteAlt != vp.inAltScreen {
		if remoteAlt {
			out.WriteString("\x1b[?1049h")
		} else {
			out.WriteString("\x1b[?1049l")
		}
		vp.inAltScreen = remoteAlt
		vp.lastCells = nil
		fullRedraw = true
	}

	if fullRedraw {
		out.WriteString("\x1b[m\x1b[2J\x1b[H")
	}

	cell := xterm.NewCellData()
	prevFg := ^uint32(0)
	prevBg := ^uint32(0)
	prevUlStyle := ^xterm.UnderlineStyle(0)
	prevUlColor := ^uint32(0)
	curRow, curCol := -1, -1

	emitMove := func(row, col int) {
		if curRow == row && curCol == col {
			return
		}
		out.WriteString(fmt.Sprintf("\x1b[%d;%dH", row+1, col+1))
		curRow, curCol = row, col
	}

	for row := 0; row < height; row++ {
		bufRow := yBase + oy + row
		var line *xterm.BufferLine
		if bufRow >= 0 && bufRow < buf.Lines.Length() {
			line = buf.Lines.Get(bufRow)
		}

		for col := 0; col < width; {
			cell.Fg = 0
			cell.Bg = 0
			cell.Extended = nil
			cell.Content = 0
			cell.CombinedData = ""

			bufCol := ox + col
			cellW := 1
			if line != nil && bufCol < vp.cols {
				line.LoadCell(bufCol, cell)
				cw := line.GetWidth(bufCol)
				if cw >= 1 {
					cellW = cw
				} else if cw == 0 {
					cellW = 1
					cell.Fg = 0
					cell.Bg = 0
					cell.Content = 0
					cell.CombinedData = ""
				}
			}

			ch := cell.GetChars()
			if ch == "" {
				ch = " "
			}
			if ch == "%" && cell.AttributeData.IsBold() != 0 && cell.AttributeData.IsInverse() != 0 {
				ch = " "
				cell.Fg = 0
				cell.Bg = 0
			}
			if cellW == 2 && col+1 >= width {
				ch = " "
				cellW = 1
				cell.Fg = 0
				cell.Bg = 0
			}

			a := &cell.AttributeData
			ulStyle := a.GetUnderlineStyle()
			var ulColor uint32
			if a.HasExtendedAttrs() != 0 && a.Extended != nil {
				ulColor = a.Extended.UnderlineColor()
			}
			newRC := renderedCell{ch: ch, fg: cell.Fg, bg: cell.Bg, width: cellW, underlineStyle: ulStyle, underlineColor: ulColor}
			if fullRedraw || vp.lastCells[row][col] != newRC {
				emitMove(row, col)
				if cell.Fg != prevFg || cell.Bg != prevBg || ulStyle != prevUlStyle || ulColor != prevUlColor {
					out.WriteString(cellAttrToSGR(cell))
					prevFg = cell.Fg
					prevBg = cell.Bg
					prevUlStyle = ulStyle
					prevUlColor = ulColor
				}
				out.WriteString(ch)
				curCol += cellW
				vp.lastCells[row][col] = newRC
				if cellW == 2 && col+1 < width {
					vp.lastCells[row][col+1] = renderedCell{width: 0}
				}
			}
			col += cellW
		}
	}

	if prevFg != 0 || prevBg != 0 || prevUlStyle != 0 || prevUlColor != 0 {
		out.WriteString("\x1b[m")
	}

	dpm := vp.term.DecPrivateModes()
	code := cursorStyleCode(dpm)
	if code != vp.lastCursorCode {
		out.WriteString(fmt.Sprintf("\x1b[%d q", code))
		vp.lastCursorCode = code
	}

	if !vp.term.IsCursorHidden() {
		localCurRow := cursorY - oy
		localCurCol := cursorX - ox
		if localCurRow >= 0 && localCurRow < height && localCurCol >= 0 && localCurCol < width {
			out.WriteString(fmt.Sprintf("\x1b[%d;%dH", localCurRow+1, localCurCol+1))
			out.WriteString("\x1b[?25h")
		}
	}

	_, _ = w.Write(out.Bytes())
}

func cellAttrToSGR(cell *xterm.CellData) string {
	a := &cell.AttributeData
	if a.IsAttributeDefault() {
		return "\x1b[m"
	}

	var sb strings.Builder
	sb.WriteString("\x1b[0")

	if a.IsBold() != 0 {
		sb.WriteString(";1")
	}
	if a.IsDim() != 0 {
		sb.WriteString(";2")
	}
	if a.IsItalic() != 0 {
		sb.WriteString(";3")
	}
	if ulStyle := a.GetUnderlineStyle(); ulStyle != xterm.UnderlineStyleNone {
		switch ulStyle {
		case xterm.UnderlineStyleDouble:
			sb.WriteString(";4:2")
		case xterm.UnderlineStyleCurly:
			sb.WriteString(";4:3")
		case xterm.UnderlineStyleDotted:
			sb.WriteString(";4:4")
		case xterm.UnderlineStyleDashed:
			sb.WriteString(";4:5")
		default:
			sb.WriteString(";4")
		}
	}
	if a.IsBlink() != 0 {
		sb.WriteString(";5")
	}
	if a.IsInverse() != 0 {
		sb.WriteString(";7")
	}
	if a.IsInvisible() != 0 {
		sb.WriteString(";8")
	}
	if a.IsStrikethrough() != 0 {
		sb.WriteString(";9")
	}
	if a.IsOverline() != 0 {
		sb.WriteString(";53")
	}

	switch a.Fg & xterm.AttrCMMask {
	case xterm.AttrCMP16:
		n := a.GetFgColor()
		if n < 8 {
			sb.WriteString(";" + strconv.Itoa(30+n))
		} else {
			sb.WriteString(";" + strconv.Itoa(90+n-8))
		}
	case xterm.AttrCMP256:
		sb.WriteString(";38;5;" + strconv.Itoa(a.GetFgColor()))
	case xterm.AttrCMRGB:
		c := xterm.ToColorRGB(uint32(a.GetFgColor()))
		sb.WriteString(fmt.Sprintf(";38;2;%d;%d;%d", c[0], c[1], c[2]))
	}

	switch a.Bg & xterm.AttrCMMask {
	case xterm.AttrCMP16:
		n := a.GetBgColor()
		if n < 8 {
			sb.WriteString(";" + strconv.Itoa(40+n))
		} else {
			sb.WriteString(";" + strconv.Itoa(100+n-8))
		}
	case xterm.AttrCMP256:
		sb.WriteString(";48;5;" + strconv.Itoa(a.GetBgColor()))
	case xterm.AttrCMRGB:
		c := xterm.ToColorRGB(uint32(a.GetBgColor()))
		sb.WriteString(fmt.Sprintf(";48;2;%d;%d;%d", c[0], c[1], c[2]))
	}

	if a.HasExtendedAttrs() != 0 && a.Extended != nil {
		uc := a.Extended.UnderlineColor()
		switch uc & xterm.AttrCMMask {
		case xterm.AttrCMP16, xterm.AttrCMP256:
			sb.WriteString(";58;5;" + strconv.Itoa(int(uc&xterm.AttrPColorMask)))
		case xterm.AttrCMRGB:
			c := xterm.ToColorRGB(uc & xterm.AttrRGBMask)
			sb.WriteString(fmt.Sprintf(";58;2;%d;%d;%d", c[0], c[1], c[2]))
		}
	}

	sb.WriteString("m")
	return sb.String()
}

func cursorStyleCode(dpm xterm.DecPrivateModes) int {
	if dpm.CursorStyle == nil && dpm.CursorBlinkOverride == nil {
		return 0
	}
	style := xterm.CursorStyleBlock
	if dpm.CursorStyle != nil {
		style = *dpm.CursorStyle
	}
	blink := true
	if dpm.CursorBlinkOverride != nil {
		blink = *dpm.CursorBlinkOverride
	} else if dpm.CursorBlink != nil {
		blink = *dpm.CursorBlink
	}
	switch style {
	case xterm.CursorStyleBlock:
		if blink {
			return 1
		}
		return 2
	case xterm.CursorStyleUnderline:
		if blink {
			return 3
		}
		return 4
	case xterm.CursorStyleBar:
		if blink {
			return 5
		}
		return 6
	}
	return 0
}
