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
