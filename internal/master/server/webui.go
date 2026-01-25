package server

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"gimpel/web" 
)

func webUIHandler() http.Handler {