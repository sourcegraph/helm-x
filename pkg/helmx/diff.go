package helmx

import (
	"fmt"
	"io"
	"os/exec"
)

type DiffOpts struct {
	*ChartifyOpts
	*ClientOpts

	Chart string

	kubeConfig string

	Out io.Writer
}

func (o DiffOpts) GetSetValues() []string {
	return o.SetValues
}

func (o DiffOpts) GetValuesFiles() []string {
	return o.ValuesFiles
}

func (o DiffOpts) GetNamespace() string {
	return o.Namespace
}

func (o DiffOpts) GetKubeContext() string {
	return o.KubeContext
}

func (o DiffOpts) GetTLS() bool {
	return o.TLS
}

func (o DiffOpts) GetTLSCert() string {
	return o.TLSCert
}

func (o DiffOpts) GetTLSKey() string {
	return o.TLSKey
}

type DiffOptionsProvider interface {
	GetSetValues() []string
	GetValuesFiles() []string
	GetNamespace() string
	GetKubeContext() string
	GetTLS() bool
	GetTLSCert() string
	GetTLSKey() string
}

type DiffOption func(*DiffOpts) error

func WithDiffOptions(opts DiffOptionsProvider) DiffOption {
	return func(o *DiffOpts) error {
		o.SetValues = opts.GetSetValues()
		o.ValuesFiles = opts.GetValuesFiles()
		o.Namespace = opts.GetNamespace()
		o.KubeContext = opts.GetKubeContext()
		o.TLS = opts.GetTLS()
		o.TLSCert = opts.GetTLSCert()
		o.TLSKey = opts.GetTLSKey()
		return nil
	}
}

// Diff returns true when the diff succeeds and changes are detected.
func (r *Runner) Diff(release, chart string, opts ...DiffOption) (bool, error) {
	o := &DiffOpts{}

	for i := range opts {
		if err := opts[i](o); err != nil {
			return false, err
		}
	}

	var additionalFlags string
	additionalFlags += createFlagChain("set", o.GetSetValues())
	additionalFlags += createFlagChain("f", o.GetValuesFiles())
	additionalFlags += createFlagChain("allow-unreleased", []string{""})
	additionalFlags += createFlagChain("detailed-exitcode", []string{""})
	additionalFlags += createFlagChain("context", []string{"3"})
	additionalFlags += createFlagChain("reset-values", []string{""})
	additionalFlags += createFlagChain("suppress-secrets", []string{""})
	if o.GetNamespace() != "" {
		additionalFlags += createFlagChain("namespace", []string{o.GetNamespace()})
	}
	if o.GetKubeContext() != "" {
		additionalFlags += createFlagChain("kube-context", []string{o.GetKubeContext()})
	}
	if o.GetTLS() {
		additionalFlags += createFlagChain("tls", []string{""})
	}
	if o.GetTLSCert() != "" {
		additionalFlags += createFlagChain("tls-cert", []string{o.GetTLSCert()})
	}
	if o.GetTLSKey() != "" {
		additionalFlags += createFlagChain("tls-key", []string{o.GetTLSKey()})
	}

	command := fmt.Sprintf("helm diff upgrade %s %s%s", release, chart, additionalFlags)
	if err := r.DeprecatedExec(command); err != nil {
		switch e := err.(type) {
		case *exec.ExitError:
			if e.ExitCode() == 2 {
				return true, nil
			}
		}
		return false, err

	}
	return false, nil
}
