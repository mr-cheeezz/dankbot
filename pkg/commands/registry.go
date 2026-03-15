package commands

import (
	"sort"
	"strings"
)

type Definition struct {
	Name           string
	Module         string
	Description    string
	Usage          string
	Example        string
	CanDisable     bool
	Configurable   bool
	DefaultEnabled bool
}

type entry struct {
	handler    Handler
	definition Definition
}

type Registry struct {
	entries map[string]entry
}

func NewRegistry() *Registry {
	return &Registry{
		entries: make(map[string]entry),
	}
}

func (r *Registry) Register(name string, handler Handler) {
	r.RegisterWithDefinition(name, handler, Definition{})
}

func (r *Registry) RegisterWithDefinition(name string, handler Handler, definition Definition) {
	if name == "" || handler == nil {
		return
	}

	name = strings.ToLower(name)
	definition.Name = name
	r.entries[name] = entry{
		handler:    handler,
		definition: definition,
	}
}

func (r *Registry) Lookup(name string) (Handler, bool) {
	entry, ok := r.entries[strings.ToLower(name)]
	return entry.handler, ok
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.entries))
	for name := range r.entries {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *Registry) Definitions() []Definition {
	names := r.Names()
	definitions := make([]Definition, 0, len(names))
	for _, name := range names {
		definitions = append(definitions, r.entries[name].definition)
	}

	return definitions
}
