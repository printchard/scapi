package spec

import (
	"fmt"
	"strings"
)

type Formatter struct {
	indent  string
	builder strings.Builder
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

func (p *Formatter) String() string {
	return p.builder.String()
}

func NewFormatter() *Formatter {
	return &Formatter{indent: "", builder: strings.Builder{}}
}
