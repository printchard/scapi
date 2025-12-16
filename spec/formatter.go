package spec

import (
	"fmt"
	"strings"
)

type Formatter struct {
	indent      string
	builder     strings.Builder
	lineBuilder strings.Builder
}

func (p *Formatter) Indent() {
	p.indent += "  "
}

func (p *Formatter) Dedent() {
	if len(p.indent) >= 2 {
		p.indent = p.indent[:len(p.indent)-2]
	}
}

func (p *Formatter) Line(format string, args ...any) {
	p.builder.WriteString(fmt.Sprintf("%s%s\n", p.indent, fmt.Sprintf(format, args...)))
}

func (p *Formatter) Partial(format string, args ...any) {
	p.lineBuilder.WriteString(fmt.Sprintf(format, args...))
}

func (p *Formatter) Flush() {
	if p.lineBuilder.Len() > 0 {
		p.builder.WriteString(fmt.Sprintf("%s%s", p.indent, p.lineBuilder.String()))
		p.lineBuilder.Reset()
	}
}

func (p *Formatter) String() string {
	return p.builder.String()
}

func NewFormatter() *Formatter {
	return &Formatter{indent: "", builder: strings.Builder{}}
}
