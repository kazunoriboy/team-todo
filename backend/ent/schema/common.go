package schema

import "regexp"

// slugRegex validates URL-friendly slugs (lowercase letters, numbers, and hyphens)
var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

