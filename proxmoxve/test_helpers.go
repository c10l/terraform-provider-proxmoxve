package proxmoxve

import (
	"fmt"
	"regexp"
)

func testAccRegexpMatch(regex string) func(string) error {
	return func(v string) error {
		match, err := regexp.Match(regex, []byte(v))
		if err != nil {
			return err
		}
		if match {
			return nil
		}
		return fmt.Errorf("expected %s, got %s", regex, v)
	}
}
