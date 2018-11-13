package storage

import "testing"

func testIsInvalidName(name string, expected bool) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()
		t.Parallel()
		actual := isInvalidName(name)

		if expected != actual {
			t.Errorf("name was \"%s\" expected %t was %t", name, expected, actual)
		}
	}

}

func TestIsInvalidName(t *testing.T) {
	t.Run("(empty)", testIsInvalidName("", true))
	t.Run("registry", testIsInvalidName("registry", false))
	t.Run("Registry", testIsInvalidName("Registry", false))
	t.Run("Bad choice", testIsInvalidName("Bad choice", true))
	t.Run("slow-registry", testIsInvalidName("slow-registry", false))
	t.Run("slow_registry", testIsInvalidName("slow_registry", false))
	t.Run("registry1", testIsInvalidName("registry1", false))
	t.Run("registry+1", testIsInvalidName("registry+1", true))
	t.Run("good-Choice", testIsInvalidName("goodChoice", false))

}
