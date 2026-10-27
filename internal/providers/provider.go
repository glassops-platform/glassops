// Package providers holds the seam between the neutral core and each delivery target type.
//
// Phase 1 defines only what doctor and init need: detection and the tools a provider relies
// on. Authenticate, Describe, Plan, Validate and Deploy join the interface in later phases,
// when there is code that implements them.
package providers

// Tool is an external program a provider needs on the PATH.
type Tool struct {
	// Name is how doctor reports the tool.
	Name string
	// Version runs the tool and returns its version, or an error when it is missing or broken.
	Version func() (string, error)
}

// Provider is one kind of delivery target, such as Salesforce.
type Provider interface {
	// Name is the value used for provider in glassops.yaml.
	Name() string
	// Detect reports whether dir holds a project for this provider.
	Detect(dir string) (bool, error)
	// Tools lists the external programs the provider needs.
	Tools() []Tool
}

// Registry lists every known provider in detection order.
type Registry []Provider

// Detect returns the first provider that recognises dir, or nil when none does.
func (r Registry) Detect(dir string) (Provider, error) {
	for _, p := range r {
		ok, err := p.Detect(dir)
		if err != nil {
			return nil, err
		}
		if ok {
			return p, nil
		}
	}
	return nil, nil
}

// Get returns the provider with the given name, or nil.
func (r Registry) Get(name string) Provider {
	for _, p := range r {
		if p.Name() == name {
			return p
		}
	}
	return nil
}
