package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	playground "github.com/go-playground/validator/v10"
	"github.com/wyw14/cry-047/internal/domain"
)

type Validator struct{ inner *playground.Validate }

func New() *Validator {
	inner := playground.New(playground.WithRequiredStructEnabled())
	inner.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return &Validator{inner: inner}
}

func (v *Validator) Struct(value any) error {
	if err := v.inner.Struct(value); err != nil {
		violations := make([]domain.FieldViolation, 0)
		var fields playground.ValidationErrors
		if errors.As(err, &fields) {
			for _, field := range fields {
				violations = append(violations, domain.FieldViolation{Field: field.Field(), Message: fmt.Sprintf("不满足规则 %s", field.Tag())})
			}
			return &domain.ValidationError{Message: "请求参数校验失败", Violations: violations}
		}
		return err
	}
	return nil
}
