// Package deps holds temporary side-effect imports so required dependencies stay in the module until wiring lands.
package deps

import (
	_ "github.com/gorilla/handlers"
	_ "github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)
