package spec

type PathTemplate struct {
	template string
	params   []string
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
