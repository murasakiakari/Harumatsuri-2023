package utility

import "strings"

type Errors []error

func (errs Errors) Error() string {
	builder := strings.Builder{}
	for _, err := range errs {
		builder.WriteString(err.Error())
		builder.WriteString("\n")
	}
	return builder.String()
}

func (errs Errors) HasError() bool {
	return len(errs) != 0
}
