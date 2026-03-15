package commands

import "strings"

type Dispatcher struct {
	prefix   string
	registry *Registry
}

func NewDispatcher(prefix string) *Dispatcher {
	return &Dispatcher{
		prefix:   prefix,
		registry: NewRegistry(),
	}
}

func (d *Dispatcher) Register(name string, handler Handler) {
	d.registry.Register(name, handler)
}

func (d *Dispatcher) RegisterWithDefinition(name string, handler Handler, definition Definition) {
	d.registry.RegisterWithDefinition(name, handler, definition)
}

func (d *Dispatcher) Names() []string {
	return d.registry.Names()
}

func (d *Dispatcher) Definitions() []Definition {
	return d.registry.Definitions()
}

func (d *Dispatcher) Dispatch(ctx Context) (Result, bool, error) {
	if !strings.HasPrefix(ctx.Message, d.prefix) {
		return Result{}, false, nil
	}

	commandLine := strings.TrimSpace(strings.TrimPrefix(ctx.Message, d.prefix))
	if commandLine == "" {
		return Result{}, false, nil
	}

	parts := strings.Fields(commandLine)
	if len(parts) == 0 {
		return Result{}, false, nil
	}

	commandName := strings.ToLower(parts[0])
	argsOffset := 1

	if len(parts) > 1 {
		twoWord := strings.ToLower(parts[0] + " " + parts[1])
		if _, ok := d.registry.Lookup(twoWord); ok {
			commandName = twoWord
			argsOffset = 2
		}
	}

	handler, ok := d.registry.Lookup(commandName)
	if !ok {
		return Result{}, false, nil
	}

	ctx.Command = commandName
	if len(parts) > argsOffset {
		ctx.Args = append([]string(nil), parts[argsOffset:]...)
	}

	result, err := handler(ctx)
	return result, true, err
}
