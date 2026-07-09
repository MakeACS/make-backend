package scalars

import (
	"fmt"
	"io"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

func MarshalClockTime(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, fmt.Sprintf(`"%s"`, t.Format("15:04:05")))
	})
}

func UnmarshalClockTime(v any) (time.Time, error) {
	str, ok := v.(string)
	if !ok {
		return time.Time{}, fmt.Errorf("ClockTimes must be strings")
	}

	return time.Parse("15:04:05", str)
}
