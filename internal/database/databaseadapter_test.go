package database

import (
	"flag"
	"fmt"
	"testing"
)

func TestMain(m *testing.M) {
	fmt.Println(flag.Lookup("test.v"))
}

func TestSiteTableCallGetAll(t *testing.T) {
	fmt.Println("Tets")
	t.Error()
}
