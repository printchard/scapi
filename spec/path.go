package spec

type PathTemplate struct {
	template string
	params   []string
}

type PathComponent struct {
	Literal string
	IsParam bool
}

func (path *PathTemplate) String() string {
	return path.template
}

func (path *PathTemplate) Params() []string {
	return path.params
}

func (path *PathTemplate) FormatString() string {
	result := ""
	for i := 0; i < len(path.template); i++ {
		if path.template[i] == '{' {
			j := i + 1
			for j < len(path.template) && path.template[j] != '}' {
				j++
			}
			result += "%s"
			i = j
		} else {
			result += string(path.template[i])
		}

	}
	return result
}

func (path *PathTemplate) Components() []PathComponent {
	var components []PathComponent
	var currentLiteral string
	for i := 0; i < len(path.template); i++ {
		if path.template[i] == '{' {
			if currentLiteral != "" {
				components = append(components, PathComponent{
					Literal: currentLiteral,
					IsParam: false,
				})
				currentLiteral = ""
			}
			param := ""
			i++
			for i < len(path.template) && path.template[i] != '}' {
				param += string(path.template[i])
				i++
			}
			components = append(components, PathComponent{
				Literal: param,
				IsParam: true,
			})
		} else {
			currentLiteral += string(path.template[i])
		}
	}
	if currentLiteral != "" {
		components = append(components, PathComponent{
			Literal: currentLiteral,
			IsParam: false,
		})
	}
	return components
}

func NewPathTemplate(template string) *PathTemplate {
	var params []string
	var param string
	for i := 0; i < len(template); i++ {
		if template[i] == '{' {
			param = ""
			i++
			for i < len(template) && template[i] != '}' {
				param += string(template[i])
				i++
			}
			params = append(params, param)
		}
	}
	return &PathTemplate{
		template: template,
		params:   params,
	}
}
