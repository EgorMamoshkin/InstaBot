package er

import "fmt"

func Wrap(msg string, err error) error {
	fmt.Errorf("%s: %w", msg, err)
}
