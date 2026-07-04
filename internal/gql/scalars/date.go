package scalars

import (
	"fmt"
	"io"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

func MarshalDate(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, fmt.Sprintf(`"%s"`, t.Format("2006-01-02")))
	})
}

func UnmarshalDate(v any) (time.Time, error) {
	str, ok := v.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("Dates must be strings")
	}

	return time.Parse("2006-01-02", str)
}
